package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrCustomerNotFound                     = errors.New("customer not found")
	ErrCustomerNotInPool                    = errors.New("customer not in pool")
	ErrCustomerAlreadyInPool                = errors.New("customer already in pool")
	ErrCustomerNotOwned                     = errors.New("customer not owned")
	ErrCustomerConvertForbidden             = errors.New("customer convert forbidden")
	ErrCustomerNameExists                   = errors.New("customer name already exists")
	ErrCustomerWeixinExists                 = errors.New("customer weixin already exists")
	ErrCustomerPhoneExists                  = errors.New("customer phone already exists")
	ErrCustomerNameRequired                 = errors.New("customer name is required")
	ErrCustomerLegalNameRequired            = errors.New("customer legal name is required")
	ErrCustomerContactNameRequired          = errors.New("customer contact name is required")
	ErrCustomerLegalNameTooShort            = errors.New("customer legal name is too short")
	ErrCustomerContactNameTooShort          = errors.New("customer contact name is too short")
	ErrCustomerLimitExceeded                = errors.New("customer limit exceeded")
	ErrCustomerSameDepartmentClaimForbidden = errors.New("same department customer cannot be claimed")
	ErrCustomerNoOutsideSalesAvailable      = errors.New("no outside sales available")
	ErrPhoneNotFound                        = errors.New("phone not found")
	ErrPhoneAlreadyExists                   = errors.New("phone already exists for this customer")
	ErrInvalidPhoneFormat                   = errors.New("invalid phone format")
)

var (
	customerMobilePhoneRegex   = regexp.MustCompile(`^1[3-9]\d{9}$`)
	customerLandlinePhoneRegex = regexp.MustCompile(`^(?:0\d{2,3}\d{7,8}|400\d{7}|800\d{7})$`)
)

const (
	customerLimitSettingKey     = "customer_limit"
	defaultCustomerLimitSetting = 100
	claimFreezeSettingKey       = "claim_freeze_days"
	defaultClaimFreezeDays      = 7
	roleSalesInside             = "sales_inside"
	roleSalesOutside            = "sales_outside"
)

var partnerSalesOwnerRoleNames = []string{
	"sales_director", "sales_manager", "sales_staff", roleSalesInside, roleSalesOutside,
	"销售总监", "销售经理", "销售员工", "销售", "Inside销售", "Outside销售",
}

var partnerOperationRoleNames = []string{
	"ops_manager", "operation_manager", "ops_staff", "operation_staff",
	"运营经理", "运营员工", "运营",
}

type CustomerService interface {
	ListCustomers(ctx context.Context, filter model.CustomerListFilter) (model.CustomerListResult, error)
	ListCustomerAssignments(ctx context.Context, filter model.CustomerAssignmentListFilter) (model.CustomerAssignmentListResult, error)
	CreateCustomer(ctx context.Context, input model.CustomerCreateInput) (*model.Customer, error)
	UpdateCustomer(ctx context.Context, customerID int64, input model.CustomerUpdateInput) (*model.Customer, error)
	CheckUnique(ctx context.Context, input model.CustomerUniqueCheckInput) (model.CustomerUniqueCheckResult, error)
	ClaimCustomer(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error)
	ReleaseCustomer(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error)
	TransferCustomer(ctx context.Context, input model.CustomerTransferInput) (*model.Customer, error)
	ReassignCustomersByYesterdayRanking(ctx context.Context, input model.CustomerBatchRankedReassignInput) (model.CustomerBatchRankedReassignResult, error)
	ConvertCustomer(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error)

	// Phone management
	AddPhone(ctx context.Context, phone *model.CustomerPhone) error
	ListPhones(ctx context.Context, customerID int64) ([]model.CustomerPhone, error)
	UpdatePhone(ctx context.Context, phone *model.CustomerPhone) error
	DeletePhone(ctx context.Context, customerID, phoneID int64) error

	// Status log management
	CreateStatusLog(ctx context.Context, log *model.CustomerStatusLog) error
	ListStatusLogs(ctx context.Context, customerID int64, page, pageSize int) ([]model.CustomerStatusLog, error)
}

type CustomerClaimFreezeError struct {
	FreezeDays  int
	Remaining   time.Duration
	FrozenUntil time.Time
	BlockType   string
}

func (e *CustomerClaimFreezeError) Error() string {
	return "customer claim is frozen"
}

const (
	customerClaimFreezeBlockTypeSelf       = "self"
	customerClaimFreezeBlockTypeDepartment = "department"
)

type customerService struct {
	repo            repository.CustomerRepository
	settingReader   customerSettingReader
	activityLogRepo *repository.ActivityLogRepository
}

type customerSettingReader interface {
	GetSetting(key string) (*model.SystemSetting, error)
}

func NewCustomerService(repo repository.CustomerRepository, settingReader customerSettingReader, activityLogRepo ...*repository.ActivityLogRepository) CustomerService {
	svc := &customerService{
		repo:          repo,
		settingReader: settingReader,
	}
	if len(activityLogRepo) > 0 {
		svc.activityLogRepo = activityLogRepo[0]
	}
	return svc
}

func (s *customerService) ListCustomers(ctx context.Context, filter model.CustomerListFilter) (model.CustomerListResult, error) {
	scoped, err := s.applyCustomerListScopeByRole(ctx, filter)
	if err != nil {
		return model.CustomerListResult{}, err
	}
	return s.repo.List(ctx, scoped)
}

func (s *customerService) ListCustomerAssignments(ctx context.Context, filter model.CustomerAssignmentListFilter) (model.CustomerAssignmentListResult, error) {
	return s.repo.ListAssignments(ctx, filter)
}

func (s *customerService) applyCustomerListScopeByRole(ctx context.Context, filter model.CustomerListFilter) (model.CustomerListFilter, error) {
	switch strings.ToLower(strings.TrimSpace(filter.Category)) {
	case "my":
		return s.applyMyCustomerScopeByRole(ctx, filter)
	case "partner":
		return s.applyPartnerCustomerScopeByRole(ctx, filter)
	default:
		return filter, nil
	}
}

func (s *customerService) applyMyCustomerScopeByRole(ctx context.Context, filter model.CustomerListFilter) (model.CustomerListFilter, error) {
	if !strings.EqualFold(strings.TrimSpace(filter.Category), "my") {
		return filter, nil
	}
	if !filter.HasViewer || filter.ViewerID <= 0 {
		return filter, nil
	}

	role := strings.TrimSpace(filter.ActorRole)
	isAdmin := isRole(role, "admin", "管理员")
	if isInsideSalesRole(role) {
		filter.AllowedOwnerUserIDs = []int64{filter.ViewerID}
		filter.AllowedInsideSalesUserIDs = []int64{filter.ViewerID}
		filter.IncludePendingConvertScope = true
		return filter, nil
	}
	scope := normalizeOwnershipScope(filter.OwnershipScope)

	directSubordinateIDs, err := s.repo.ListDirectSubordinateUserIDsByRoleNames(ctx, []int64{filter.ViewerID}, nil)
	if err != nil {
		return model.CustomerListFilter{}, err
	}
	directSubordinateIDs = uniquePositiveInt64(directSubordinateIDs)

	subordinateIDs, err := s.listAllDescendantUserIDs(ctx, filter.ViewerID)
	if err != nil {
		return model.CustomerListFilter{}, err
	}
	selfAndSubordinates := uniquePositiveInt64(append([]int64{filter.ViewerID}, subordinateIDs...))

	switch scope {
	case "mine":
		filter.AllowedOwnerUserIDs = []int64{filter.ViewerID}
	case "subordinates":
		filter.AllowedOwnerUserIDs = directSubordinateIDs
	case "inside_sales":
		insideSalesIDs, err := s.repo.ListUserIDsByRoleNames(ctx, []string{
			roleSalesInside, "sale_inside", "Inside销售", "inside销售", "电销员工",
		})
		if err != nil {
			return model.CustomerListFilter{}, err
		}
		insideSalesIDs = uniquePositiveInt64(insideSalesIDs)
		filter.IncludePoolInMyScope = true
		filter.IncludePendingConvertScope = true
		filter.RequireInsideSalesAssociation = true
		if isAdmin {
			filter.AllowedInsideSalesUserIDs = insideSalesIDs
			filter.SkipViewerOwnerLimit = true
			break
		}
		filter.AllowedInsideSalesUserIDs = intersectPositiveInt64(selfAndSubordinates, insideSalesIDs)
		filter.AllowedOwnerUserIDs = selfAndSubordinates
	case "sales":
		salesIDs, err := s.repo.ListUserIDsByRoleNames(ctx, []string{
			"sales_director", "sales_manager", "sales_staff", roleSalesInside, roleSalesOutside,
			"销售总监", "销售经理", "销售员工", "销售", "Inside销售", "Outside销售",
		})
		if err != nil {
			return model.CustomerListFilter{}, err
		}
		salesIDs = uniquePositiveInt64(salesIDs)
		if isAdmin {
			filter.AllowedOwnerUserIDs = salesIDs
			break
		}
		filter.AllowedOwnerUserIDs = intersectPositiveInt64(selfAndSubordinates, salesIDs)
	case "all":
		fallthrough
	default:
		if isAdmin {
			filter.SkipViewerOwnerLimit = true
			filter.AllowedOwnerUserIDs = nil
			return filter, nil
		}
		filter.AllowedOwnerUserIDs = selfAndSubordinates
	}

	if scope == "inside_sales" {
		if len(filter.AllowedInsideSalesUserIDs) == 0 && len(filter.AllowedOwnerUserIDs) == 0 {
			filter.AllowedInsideSalesUserIDs = []int64{-1}
			filter.AllowedOwnerUserIDs = []int64{-1}
		}
		return filter, nil
	}

	if len(filter.AllowedOwnerUserIDs) == 0 {
		filter.AllowedOwnerUserIDs = []int64{-1}
	}

	return filter, nil
}

func (s *customerService) applyPartnerCustomerScopeByRole(ctx context.Context, filter model.CustomerListFilter) (model.CustomerListFilter, error) {
	if !strings.EqualFold(strings.TrimSpace(filter.Category), "partner") {
		return filter, nil
	}
	if !filter.HasViewer || filter.ViewerID <= 0 {
		return filter, nil
	}

	scope, err := s.resolvePartnerCustomerAccessScope(ctx, filter.ViewerID, filter.ActorRole)
	if err != nil {
		return model.CustomerListFilter{}, err
	}

	if len(scope.allowedOwnerUserIDs) > 0 {
		filter.AllowedOwnerUserIDs = scope.allowedOwnerUserIDs
		filter.AllowedServiceUserIDs = nil
		filter.ForceServiceUserID = nil
		return filter, nil
	}
	if len(scope.allowedServiceUserIDs) > 0 {
		filter.AllowedOwnerUserIDs = nil
		filter.AllowedServiceUserIDs = scope.allowedServiceUserIDs
		filter.ForceServiceUserID = nil
		return filter, nil
	}
	if scope.forceServiceUserID != nil {
		filter.AllowedOwnerUserIDs = nil
		filter.AllowedServiceUserIDs = nil
		filter.ForceServiceUserID = scope.forceServiceUserID
		return filter, nil
	}

	filter.AllowedOwnerUserIDs = []int64{filter.ViewerID}
	return filter, nil
}

type partnerCustomerAccessScope struct {
	allowedOwnerUserIDs   []int64
	allowedServiceUserIDs []int64
	forceServiceUserID    *int64
}

func (s *customerService) resolvePartnerCustomerAccessScope(ctx context.Context, actorUserID int64, actorRole string) (partnerCustomerAccessScope, error) {
	scope := partnerCustomerAccessScope{}
	role := strings.TrimSpace(actorRole)
	if actorUserID <= 0 {
		return scope, nil
	}

	switch {
	case isRole(role, "admin", "管理员", "finance_manager", "finance", "财务经理", "财务"):
		salesIDs, err := s.repo.ListUserIDsByRoleNames(ctx, partnerSalesOwnerRoleNames)
		if err != nil {
			return partnerCustomerAccessScope{}, err
		}
		scope.allowedOwnerUserIDs = uniquePositiveInt64(salesIDs)
		if len(scope.allowedOwnerUserIDs) == 0 {
			scope.allowedOwnerUserIDs = []int64{-1}
		}
		return scope, nil
	case isRole(role, "sales_director", "销售总监", "sales_manager", "销售经理"):
		descendantIDs, err := s.listAllDescendantUserIDs(ctx, actorUserID)
		if err != nil {
			return partnerCustomerAccessScope{}, err
		}
		salesIDs, err := s.repo.ListUserIDsByRoleNames(ctx, partnerSalesOwnerRoleNames)
		if err != nil {
			return partnerCustomerAccessScope{}, err
		}
		scope.allowedOwnerUserIDs = intersectPositiveInt64(
			uniquePositiveInt64(append([]int64{actorUserID}, descendantIDs...)),
			uniquePositiveInt64(salesIDs),
		)
		if len(scope.allowedOwnerUserIDs) == 0 {
			scope.allowedOwnerUserIDs = []int64{-1}
		}
		return scope, nil
	case isRole(role, "ops_manager", "operation_manager", "运营经理"):
		operationIDs, err := s.repo.ListUserIDsByRoleNames(ctx, partnerOperationRoleNames)
		if err != nil {
			return partnerCustomerAccessScope{}, err
		}
		scope.allowedServiceUserIDs = uniquePositiveInt64(operationIDs)
		if len(scope.allowedServiceUserIDs) == 0 {
			scope.allowedServiceUserIDs = []int64{-1}
		}
		return scope, nil
	case isRole(role, "ops_staff", "operation_staff", "运营员工", "运营"):
		force := actorUserID
		scope.forceServiceUserID = &force
		return scope, nil
	default:
		scope.allowedOwnerUserIDs = []int64{actorUserID}
		return scope, nil
	}
}

func normalizeOwnershipScope(raw string) string {
	scope := strings.ToLower(strings.TrimSpace(raw))
	switch scope {
	case "", "all", "全部", "全部客户":
		return "all"
	case "mine", "my", "self", "我的", "本人":
		return "mine"
	case "sales", "sales_department", "sales-department", "salesdept", "销售", "销售部":
		return "sales"
	case "inside_sales", "inside-sales", "inside_sales_department", "telemarketing", "电销", "电销部":
		return "inside_sales"
	case "subordinate", "subordinates", "team", "下属", "团队":
		return "subordinates"
	default:
		// Unknown values fallback to "all" to avoid accidental empty-list behavior.
		return "all"
	}
}

func isInsideSalesRole(role string) bool {
	return isRole(role, roleSalesInside, "sale_inside", "Inside销售", "inside销售", "电销员工")
}

func isOutsideSalesRole(role string) bool {
	return isRole(role, roleSalesOutside, "sale_outside", "Outside销售", "outside销售")
}

func uniquePositiveInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func intersectPositiveInt64(left, right []int64) []int64 {
	if len(left) == 0 || len(right) == 0 {
		return []int64{}
	}
	allowed := make(map[int64]struct{}, len(right))
	for _, id := range right {
		if id <= 0 {
			continue
		}
		allowed[id] = struct{}{}
	}
	if len(allowed) == 0 {
		return []int64{}
	}
	result := make([]int64, 0, len(left))
	seen := make(map[int64]struct{}, len(left))
	for _, id := range left {
		if id <= 0 {
			continue
		}
		if _, ok := allowed[id]; !ok {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func (s *customerService) listAllDescendantUserIDs(ctx context.Context, rootUserID int64) ([]int64, error) {
	return listAllDescendantUserIDsByRepo(ctx, s.repo, rootUserID)
}

func (s *customerService) CreateCustomer(ctx context.Context, input model.CustomerCreateInput) (*model.Customer, error) {
	normalized, uniqueInput, err := normalizeCreateInput(input)
	if err != nil {
		return nil, err
	}
	normalized, err = s.applyCreateOwnershipPolicy(ctx, normalized)
	if err != nil {
		return nil, err
	}

	result, err := s.repo.CheckUnique(ctx, uniqueInput)
	if err != nil {
		return nil, err
	}
	if result.NameExists {
		return nil, ErrCustomerNameExists
	}
	if result.WeixinExists {
		return nil, ErrCustomerWeixinExists
	}
	if len(result.DuplicatePhones) > 0 {
		return nil, ErrCustomerPhoneExists
	}
	if err := s.validateCustomerLimit(ctx, normalized); err != nil {
		return nil, err
	}

	customer, err := s.repo.Create(ctx, normalized)
	if err != nil {
		return nil, err
	}
	s.logActivity(ctx, normalized.OperatorUserID, model.ActionCreateCustomer, model.TargetTypeCustomer, customer.ID, customer.Name, "")
	return customer, nil
}

func (s *customerService) applyCreateOwnershipPolicy(ctx context.Context, input model.CustomerCreateInput) (model.CustomerCreateInput, error) {
	if input.OperatorUserID <= 0 {
		return input, nil
	}

	operatorRole, err := s.repo.GetUserRoleName(ctx, input.OperatorUserID)
	if err != nil {
		return model.CustomerCreateInput{}, err
	}

	if isInsideSalesRole(operatorRole) {
		ownerUserID, err := s.resolveConvertedCustomerOwnerUserID(ctx, input.OperatorUserID, input.OperatorUserID, operatorRole)
		if err != nil {
			return model.CustomerCreateInput{}, err
		}
		insideSalesUserID := input.OperatorUserID
		input.Status = model.CustomerStatusOwned
		input.InsideSalesUserID = &insideSalesUserID
		if ownerUserID > 0 {
			now := time.Now().UTC()
			input.OwnerUserID = &ownerUserID
			input.ConvertedAt = &now
		} else {
			input.OwnerUserID = &insideSalesUserID
			input.ConvertedAt = nil
		}
		return input, nil
	}
	if strings.TrimSpace(input.Status) != model.CustomerStatusOwned {
		return input, nil
	}

	switch {
	case isOutsideSalesRole(operatorRole):
		ownerUserID := input.OperatorUserID
		input.OwnerUserID = &ownerUserID
	}

	return input, nil
}

func (s *customerService) UpdateCustomer(ctx context.Context, customerID int64, input model.CustomerUpdateInput) (*model.Customer, error) {
	normalized, uniqueInput, err := normalizeUpdateInput(customerID, input)
	if err != nil {
		return nil, err
	}

	result, err := s.repo.CheckUnique(ctx, uniqueInput)
	if err != nil {
		return nil, err
	}
	if result.NameExists {
		return nil, ErrCustomerNameExists
	}
	if result.WeixinExists {
		return nil, ErrCustomerWeixinExists
	}
	if len(result.DuplicatePhones) > 0 {
		return nil, ErrCustomerPhoneExists
	}

	customer, err := s.repo.Update(ctx, customerID, normalized)
	if err != nil {
		if errors.Is(err, repository.ErrCustomerNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return customer, nil
}

func (s *customerService) CheckUnique(ctx context.Context, input model.CustomerUniqueCheckInput) (model.CustomerUniqueCheckResult, error) {
	normalized := model.CustomerUniqueCheckInput{
		ExcludeCustomerID: input.ExcludeCustomerID,
		Name:              strings.TrimSpace(input.Name),
		LegalName:         strings.TrimSpace(input.LegalName),
		ContactName:       strings.TrimSpace(input.ContactName),
		Weixin:            strings.TrimSpace(input.Weixin),
		Phones:            make([]string, 0, len(input.Phones)),
	}
	for _, phone := range input.Phones {
		normalizedPhone := normalizeCustomerPhone(phone)
		if normalizedPhone == "" {
			continue
		}
		normalized.Phones = append(normalized.Phones, normalizedPhone)
	}
	return s.repo.CheckUnique(ctx, normalized)
}

func (s *customerService) ClaimCustomer(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error) {
	customer, err := s.repo.FindByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, repository.ErrCustomerNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	if !customer.IsInPool {
		return nil, ErrCustomerNotInPool
	}
	if customer.DropUserID != nil && *customer.DropUserID > 0 {
		if *customer.DropUserID == operatorUserID {
			freezeErr, err := s.buildClaimFreezeError(customer)
			if err != nil {
				return nil, err
			}
			if freezeErr != nil {
				return nil, freezeErr
			}
		}
	}
	departmentFreezeErr, err := s.buildDepartmentClaimFreezeError(ctx, customerID, operatorUserID)
	if err != nil {
		return nil, err
	}
	if departmentFreezeErr != nil {
		return nil, departmentFreezeErr
	}

	operatorRole, err := s.repo.GetUserRoleName(ctx, operatorUserID)
	if err != nil {
		return nil, err
	}

	claimOwnerUserID := operatorUserID
	var insideSalesUserID *int64
	if isInsideSalesRole(operatorRole) {
		insideSalesUserID = &operatorUserID
		ownerUserID, err := s.resolveConvertedCustomerOwnerUserID(ctx, operatorUserID, operatorUserID, operatorRole)
		if err != nil {
			return nil, err
		}
		if ownerUserID > 0 {
			claimOwnerUserID = ownerUserID
		}
	}
	if err := s.validateCustomerLimitByOwner(ctx, claimOwnerUserID); err != nil {
		return nil, err
	}

	customer, err = s.repo.Claim(ctx, customerID, claimOwnerUserID, operatorUserID, insideSalesUserID)
	if err == nil {
		s.logActivity(ctx, operatorUserID, model.ActionClaimCustomer, model.TargetTypeCustomer, customer.ID, customer.Name, "")
		return customer, nil
	}
	if errors.Is(err, repository.ErrCustomerNotFound) {
		return nil, ErrCustomerNotFound
	}
	if errors.Is(err, repository.ErrCustomerNotInPool) {
		return nil, ErrCustomerNotInPool
	}
	return nil, err
}

func (s *customerService) buildDepartmentClaimFreezeError(ctx context.Context, customerID, operatorUserID int64) (*CustomerClaimFreezeError, error) {
	if customerID <= 0 || operatorUserID <= 0 {
		return nil, nil
	}

	operatorAnchorUserID, err := resolveSalesDirectorUserID(ctx, s.repo, operatorUserID)
	if err != nil {
		return nil, err
	}
	if operatorAnchorUserID <= 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	blockedUntil, err := s.repo.GetActiveBlockedUntilByDepartmentAnchor(ctx, customerID, operatorAnchorUserID, now)
	if err != nil {
		return nil, err
	}
	if blockedUntil == nil || !blockedUntil.After(now) {
		return nil, nil
	}

	freezeDays, err := s.getClaimFreezeDays()
	if err != nil {
		return nil, err
	}
	if freezeDays <= 0 {
		freezeDays = defaultClaimFreezeDays
	}

	return &CustomerClaimFreezeError{
		FreezeDays:  freezeDays,
		Remaining:   blockedUntil.Sub(now),
		FrozenUntil: *blockedUntil,
		BlockType:   customerClaimFreezeBlockTypeDepartment,
	}, nil
}

func (s *customerService) isSameDepartment(ctx context.Context, leftUserID, rightUserID int64) (bool, error) {
	if leftUserID <= 0 || rightUserID <= 0 {
		return false, nil
	}

	leftAnchorID, err := resolveSalesDirectorUserID(ctx, s.repo, leftUserID)
	if err != nil {
		return false, err
	}
	rightAnchorID, err := resolveSalesDirectorUserID(ctx, s.repo, rightUserID)
	if err != nil {
		return false, err
	}
	if leftAnchorID <= 0 || rightAnchorID <= 0 {
		return false, nil
	}
	return leftAnchorID == rightAnchorID, nil
}

func (s *customerService) ReleaseCustomer(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error) {
	customer, err := s.repo.Release(ctx, customerID, operatorUserID)
	if err == nil {
		s.logActivity(ctx, operatorUserID, model.ActionReleaseCustomer, model.TargetTypeCustomer, customer.ID, customer.Name, "")
		return customer, nil
	}
	if errors.Is(err, repository.ErrCustomerNotFound) {
		return nil, ErrCustomerNotFound
	}
	if errors.Is(err, repository.ErrCustomerAlreadyInPool) {
		return nil, ErrCustomerAlreadyInPool
	}
	if errors.Is(err, repository.ErrCustomerNotOwned) {
		return nil, ErrCustomerNotOwned
	}
	return nil, err
}

func (s *customerService) TransferCustomer(ctx context.Context, input model.CustomerTransferInput) (*model.Customer, error) {
	if input.ToOwnerUserID > 0 && input.ToOwnerUserID != input.OperatorUserID {
		if err := s.validateCustomerLimitByOwner(ctx, input.ToOwnerUserID); err != nil {
			return nil, err
		}
	}

	customer, err := s.repo.Transfer(ctx, input)
	if err == nil {
		s.logActivity(ctx, input.OperatorUserID, model.ActionTransferCustomer, model.TargetTypeCustomer, customer.ID, customer.Name, "")
		return customer, nil
	}
	if errors.Is(err, repository.ErrCustomerNotFound) {
		return nil, ErrCustomerNotFound
	}
	if errors.Is(err, repository.ErrCustomerNotOwned) {
		return nil, ErrCustomerNotOwned
	}
	return nil, err
}

func (s *customerService) ReassignCustomersByYesterdayRanking(
	ctx context.Context,
	input model.CustomerBatchRankedReassignInput,
) (model.CustomerBatchRankedReassignResult, error) {
	customerIDs := uniquePositiveInt64(input.CustomerIDs)
	result := model.CustomerBatchRankedReassignResult{
		Total: len(customerIDs),
		Items: make([]model.CustomerBatchRankedReassignItem, 0, len(customerIDs)),
	}
	if len(customerIDs) == 0 {
		return result, nil
	}

	type customerPlan struct {
		customer *model.Customer
		targetID int64
		message  string
	}

	referenceDate := previousAutoAssignScoreDate()

	groupedCustomers := make(map[int64][]*model.Customer)
	orderedAnchors := make([]int64, 0)
	seenAnchors := make(map[int64]struct{})

	for _, customerID := range customerIDs {
		customer, err := s.repo.FindByID(ctx, customerID)
		if err != nil {
			result.Items = append(result.Items, model.CustomerBatchRankedReassignItem{
				CustomerID: customerID,
				Success:    false,
				Message:    "客户不存在或已被删除",
			})
			result.FailedCount++
			continue
		}
		if customer.OwnerUserID == nil || *customer.OwnerUserID <= 0 {
			result.Items = append(result.Items, model.CustomerBatchRankedReassignItem{
				CustomerID:      customer.ID,
				CustomerName:    customer.Name,
				FromOwnerUserID: customer.OwnerUserID,
				Success:         false,
				Message:         "该客户当前没有负责人，无法重新分配",
			})
			result.FailedCount++
			continue
		}

		directorUserID, err := resolveSalesDirectorUserID(ctx, s.repo, *customer.OwnerUserID)
		if err != nil || directorUserID <= 0 {
			result.Items = append(result.Items, model.CustomerBatchRankedReassignItem{
				CustomerID:      customer.ID,
				CustomerName:    customer.Name,
				FromOwnerUserID: customer.OwnerUserID,
				Success:         false,
				Message:         "未找到该客户所属部门的销售负责人",
			})
			result.FailedCount++
			continue
		}

		if _, ok := seenAnchors[directorUserID]; !ok {
			seenAnchors[directorUserID] = struct{}{}
			orderedAnchors = append(orderedAnchors, directorUserID)
		}
		groupedCustomers[directorUserID] = append(groupedCustomers[directorUserID], customer)
	}

	planned := make(map[int64]customerPlan, len(customerIDs))
	for _, directorUserID := range orderedAnchors {
		customersInDept := groupedCustomers[directorUserID]
		candidateUserIDs, err := listAssignableSalesOwnerUserIDs(ctx, s.repo, directorUserID)
		if err != nil || len(candidateUserIDs) == 0 {
			for _, customer := range customersInDept {
				result.Items = append(result.Items, model.CustomerBatchRankedReassignItem{
					CustomerID:      customer.ID,
					CustomerName:    customer.Name,
					FromOwnerUserID: customer.OwnerUserID,
					Success:         false,
					Message:         "该部门暂无可分配销售",
				})
				result.FailedCount++
			}
			continue
		}

		rankedScores, err := s.repo.ListAutoAssignRankedOwnerScores(ctx, referenceDate, candidateUserIDs)
		if err != nil {
			for _, customer := range customersInDept {
				result.Items = append(result.Items, model.CustomerBatchRankedReassignItem{
					CustomerID:      customer.ID,
					CustomerName:    customer.Name,
					FromOwnerUserID: customer.OwnerUserID,
					Success:         false,
					Message:         "加载昨日排名失败",
				})
				result.FailedCount++
			}
			continue
		}

		eligibleOwnerUserIDs := filterAutoAssignEligibleOwnerUserIDs(rankedScores)
		if len(eligibleOwnerUserIDs) == 0 {
			for _, customer := range customersInDept {
				result.Items = append(result.Items, model.CustomerBatchRankedReassignItem{
					CustomerID:      customer.ID,
					CustomerName:    customer.Name,
					FromOwnerUserID: customer.OwnerUserID,
					Success:         false,
					Message:         "该部门昨日暂无符合规则的排名人员",
				})
				result.FailedCount++
			}
			continue
		}

		for idx, customer := range customersInDept {
			targetID := eligibleOwnerUserIDs[idx%len(eligibleOwnerUserIDs)]
			message := "已按昨日排名重新分配"
			if customer.OwnerUserID != nil && *customer.OwnerUserID == targetID {
				message = "按昨日排名计算后负责人未变化"
			}
			planned[customer.ID] = customerPlan{
				customer: customer,
				targetID: targetID,
				message:  message,
			}
		}
	}

	for _, customerID := range customerIDs {
		plan, ok := planned[customerID]
		if !ok {
			continue
		}

		item := model.CustomerBatchRankedReassignItem{
			CustomerID:      plan.customer.ID,
			CustomerName:    plan.customer.Name,
			FromOwnerUserID: plan.customer.OwnerUserID,
			Success:         true,
			Message:         plan.message,
		}
		if plan.targetID > 0 {
			item.ToOwnerUserID = &plan.targetID
		}

		if plan.customer.OwnerUserID != nil && *plan.customer.OwnerUserID == plan.targetID {
			result.Items = append(result.Items, item)
			result.SuccessCount++
			continue
		}

		_, err := s.TransferCustomer(ctx, model.CustomerTransferInput{
			CustomerID:     plan.customer.ID,
			ToOwnerUserID:  plan.targetID,
			OperatorUserID: input.OperatorUserID,
			Note:           "按昨日排名重新分配客户",
			AllowAnyOwner:  true,
		})
		if err != nil {
			item.Success = false
			item.Message = err.Error()
			item.ToOwnerUserID = nil
			result.Items = append(result.Items, item)
			result.FailedCount++
			continue
		}

		result.Items = append(result.Items, item)
		result.SuccessCount++
	}

	return result, nil
}

func (s *customerService) ConvertCustomer(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error) {
	customer, err := s.repo.FindByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, repository.ErrCustomerNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	operatorRole, err := s.repo.GetUserRoleName(ctx, operatorUserID)
	if err != nil {
		return nil, err
	}
	isAdmin := isRole(operatorRole, "admin", "管理员")

	associatedInsideSalesUserID := int64(0)
	if customer.InsideSalesUserID != nil && *customer.InsideSalesUserID > 0 {
		associatedInsideSalesUserID = *customer.InsideSalesUserID
	} else if customer.OwnerUserID != nil && *customer.OwnerUserID > 0 {
		ownerRole, roleErr := s.repo.GetUserRoleName(ctx, *customer.OwnerUserID)
		if roleErr != nil {
			return nil, roleErr
		}
		if isInsideSalesRole(ownerRole) {
			associatedInsideSalesUserID = *customer.OwnerUserID
		}
	}
	if associatedInsideSalesUserID <= 0 && customer.CreateUserID > 0 {
		associatedInsideSalesUserID = customer.CreateUserID
	}
	if associatedInsideSalesUserID <= 0 {
		return nil, ErrCustomerConvertForbidden
	}
	if !isAdmin && associatedInsideSalesUserID != operatorUserID {
		return nil, ErrCustomerConvertForbidden
	}
	if customer.ConvertedAt != nil {
		return nil, ErrCustomerConvertForbidden
	}
	if len(customer.HistoricalOwnerIDs) > 0 && customer.ConvertedAt == nil {
		return nil, ErrCustomerConvertForbidden
	}
	if customer.OwnerUserID != nil && *customer.OwnerUserID > 0 && *customer.OwnerUserID != associatedInsideSalesUserID {
		return nil, ErrCustomerConvertForbidden
	}

	ownerUserID, err := s.resolveConvertedCustomerOwnerUserID(ctx, operatorUserID, associatedInsideSalesUserID, operatorRole)
	if err != nil {
		return nil, err
	}
	if ownerUserID <= 0 {
		ownerUserID = associatedInsideSalesUserID
	}
	if err := s.validateCustomerLimitByOwner(ctx, ownerUserID); err != nil {
		return nil, err
	}

	customer, err = s.repo.Convert(ctx, customerID, ownerUserID, operatorUserID)
	if err == nil {
		s.logActivity(ctx, operatorUserID, model.ActionTransferCustomer, model.TargetTypeCustomer, customer.ID, customer.Name, "")
		return customer, nil
	}
	if errors.Is(err, repository.ErrCustomerNotFound) {
		return nil, ErrCustomerNotFound
	}
	if errors.Is(err, repository.ErrCustomerNotInPool) {
		return nil, ErrCustomerNotInPool
	}
	return nil, err
}

func (s *customerService) AddPhone(ctx context.Context, phone *model.CustomerPhone) error {
	phone.Phone = normalizeCustomerPhone(phone.Phone)
	phone.PhoneLabel = strings.TrimSpace(phone.PhoneLabel)

	if !isValidCustomerPhone(phone.Phone) {
		return ErrInvalidPhoneFormat
	}

	err := s.repo.AddPhone(ctx, phone)
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrPhoneAlreadyExists) {
		return ErrPhoneAlreadyExists
	}
	return err
}

func (s *customerService) ListPhones(ctx context.Context, customerID int64) ([]model.CustomerPhone, error) {
	return s.repo.ListPhones(ctx, customerID)
}

func (s *customerService) UpdatePhone(ctx context.Context, phone *model.CustomerPhone) error {
	phone.Phone = normalizeCustomerPhone(phone.Phone)
	phone.PhoneLabel = strings.TrimSpace(phone.PhoneLabel)

	if !isValidCustomerPhone(phone.Phone) {
		return ErrInvalidPhoneFormat
	}

	err := s.repo.UpdatePhone(ctx, phone)
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrPhoneNotFound) {
		return ErrPhoneNotFound
	}
	if errors.Is(err, repository.ErrPhoneAlreadyExists) {
		return ErrPhoneAlreadyExists
	}
	return err
}

func (s *customerService) DeletePhone(ctx context.Context, customerID, phoneID int64) error {
	err := s.repo.DeletePhone(ctx, customerID, phoneID)
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrPhoneNotFound) {
		return ErrPhoneNotFound
	}
	return err
}

func (s *customerService) CreateStatusLog(ctx context.Context, log *model.CustomerStatusLog) error {
	return s.repo.CreateStatusLog(ctx, log)
}

func (s *customerService) ListStatusLogs(ctx context.Context, customerID int64, page, pageSize int) ([]model.CustomerStatusLog, error) {
	return s.repo.ListStatusLogs(ctx, customerID, page, pageSize)
}

func normalizeCreateInput(input model.CustomerCreateInput) (model.CustomerCreateInput, model.CustomerUniqueCheckInput, error) {
	normalizedPhones, err := normalizePhones(input.Phones)
	if err != nil {
		return model.CustomerCreateInput{}, model.CustomerUniqueCheckInput{}, err
	}

	normalized := model.CustomerCreateInput{
		Name:           strings.TrimSpace(input.Name),
		LegalName:      strings.TrimSpace(input.LegalName),
		ContactName:    strings.TrimSpace(input.ContactName),
		Weixin:         strings.TrimSpace(input.Weixin),
		Email:          strings.TrimSpace(input.Email),
		Province:       input.Province,
		City:           input.City,
		Area:           input.Area,
		DetailAddress:  strings.TrimSpace(input.DetailAddress),
		Remark:         strings.TrimSpace(input.Remark),
		Status:         strings.TrimSpace(input.Status),
		OwnerUserID:    input.OwnerUserID,
		OperatorUserID: input.OperatorUserID,
		Phones:         normalizedPhones,
	}

	if normalized.Name == "" {
		return model.CustomerCreateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerNameRequired
	}
	if normalized.LegalName == "" {
		return model.CustomerCreateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerLegalNameRequired
	}
	if normalized.ContactName == "" {
		return model.CustomerCreateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerContactNameRequired
	}
	if len([]rune(normalized.LegalName)) < 2 {
		return model.CustomerCreateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerLegalNameTooShort
	}
	if len([]rune(normalized.ContactName)) < 2 {
		return model.CustomerCreateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerContactNameTooShort
	}

	phoneValues := make([]string, 0, len(normalizedPhones))
	for _, phone := range normalizedPhones {
		phoneValues = append(phoneValues, phone.Phone)
	}

	return normalized, model.CustomerUniqueCheckInput{
		Name:        normalized.Name,
		LegalName:   normalized.LegalName,
		ContactName: normalized.ContactName,
		Weixin:      normalized.Weixin,
		Phones:      phoneValues,
	}, nil
}

func (s *customerService) validateCustomerLimit(ctx context.Context, input model.CustomerCreateInput) error {
	ownerUserID, ok := resolveOwnedCustomerOwner(input)
	if !ok {
		return nil
	}
	return s.validateCustomerLimitByOwner(ctx, ownerUserID)
}

func (s *customerService) validateCustomerLimitByOwner(ctx context.Context, ownerUserID int64) error {
	if ownerUserID <= 0 {
		return nil
	}

	limit, err := s.getCustomerLimit()
	if err != nil {
		return err
	}
	if limit <= 0 {
		return nil
	}

	total, err := s.repo.CountOwnedActiveByOwner(ctx, ownerUserID)
	if err != nil {
		return err
	}
	if total >= int64(limit) {
		return ErrCustomerLimitExceeded
	}
	return nil
}

func (s *customerService) buildClaimFreezeError(customer *model.Customer) (*CustomerClaimFreezeError, error) {
	if customer == nil || customer.DropTime == nil {
		return nil, nil
	}

	freezeDays, err := s.getClaimFreezeDays()
	if err != nil {
		return nil, err
	}
	if freezeDays <= 0 {
		return nil, nil
	}

	now := time.Now().UTC()
	frozenUntil := customer.DropTime.Add(time.Duration(freezeDays) * 24 * time.Hour)
	remaining := frozenUntil.Sub(now)
	if remaining <= 0 {
		return nil, nil
	}

	return &CustomerClaimFreezeError{
		FreezeDays:  freezeDays,
		Remaining:   remaining,
		FrozenUntil: frozenUntil,
		BlockType:   customerClaimFreezeBlockTypeSelf,
	}, nil
}

func (s *customerService) getCustomerLimit() (int, error) {
	if s.settingReader == nil {
		return defaultCustomerLimitSetting, nil
	}

	setting, err := s.settingReader.GetSetting(customerLimitSettingKey)
	if err != nil {
		return 0, err
	}
	if setting == nil {
		return defaultCustomerLimitSetting, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(setting.Value))
	if err != nil {
		return defaultCustomerLimitSetting, nil
	}
	return value, nil
}

func (s *customerService) getClaimFreezeDays() (int, error) {
	if s.settingReader == nil {
		return defaultClaimFreezeDays, nil
	}

	setting, err := s.settingReader.GetSetting(claimFreezeSettingKey)
	if err != nil {
		return 0, err
	}
	if setting == nil {
		return defaultClaimFreezeDays, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(setting.Value))
	if err != nil {
		return defaultClaimFreezeDays, nil
	}
	if value < 0 {
		return 0, nil
	}
	return value, nil
}

func resolveOwnedCustomerOwner(input model.CustomerCreateInput) (int64, bool) {
	if strings.TrimSpace(input.Status) != model.CustomerStatusOwned {
		return 0, false
	}
	if input.OwnerUserID != nil && *input.OwnerUserID > 0 {
		return *input.OwnerUserID, true
	}
	if input.OperatorUserID > 0 {
		return input.OperatorUserID, true
	}
	return 0, false
}

func (s *customerService) resolveConvertedCustomerOwnerUserID(ctx context.Context, operatorUserID, createUserID int64, operatorRole string) (int64, error) {
	if operatorUserID <= 0 {
		return 0, ErrCustomerConvertForbidden
	}

	assignmentUserID := operatorUserID
	assignmentRole := operatorRole
	if isRole(operatorRole, "admin", "管理员") && createUserID > 0 {
		assignmentUserID = createUserID
		roleName, err := s.repo.GetUserRoleName(ctx, createUserID)
		if err != nil {
			return 0, err
		}
		assignmentRole = roleName
	}

	switch {
	case isInsideSalesRole(assignmentRole):
		return pickBalancedSalesOwnerUserID(ctx, s.repo, assignmentUserID)
	case isOutsideSalesRole(assignmentRole):
		return assignmentUserID, nil
	default:
		return 0, ErrCustomerConvertForbidden
	}
}

func normalizeUpdateInput(customerID int64, input model.CustomerUpdateInput) (model.CustomerUpdateInput, model.CustomerUniqueCheckInput, error) {
	normalizedPhones, err := normalizePhones(input.Phones)
	if err != nil {
		return model.CustomerUpdateInput{}, model.CustomerUniqueCheckInput{}, err
	}

	normalized := model.CustomerUpdateInput{
		Name:           strings.TrimSpace(input.Name),
		LegalName:      strings.TrimSpace(input.LegalName),
		ContactName:    strings.TrimSpace(input.ContactName),
		Weixin:         strings.TrimSpace(input.Weixin),
		Email:          strings.TrimSpace(input.Email),
		Province:       input.Province,
		City:           input.City,
		Area:           input.Area,
		DetailAddress:  strings.TrimSpace(input.DetailAddress),
		Remark:         strings.TrimSpace(input.Remark),
		OperatorUserID: input.OperatorUserID,
		Phones:         normalizedPhones,
	}

	if normalized.Name == "" {
		return model.CustomerUpdateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerNameRequired
	}
	if normalized.LegalName == "" {
		return model.CustomerUpdateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerLegalNameRequired
	}
	if normalized.ContactName == "" {
		return model.CustomerUpdateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerContactNameRequired
	}
	if len([]rune(normalized.LegalName)) < 2 {
		return model.CustomerUpdateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerLegalNameTooShort
	}
	if len([]rune(normalized.ContactName)) < 2 {
		return model.CustomerUpdateInput{}, model.CustomerUniqueCheckInput{}, ErrCustomerContactNameTooShort
	}

	phoneValues := make([]string, 0, len(normalizedPhones))
	for _, phone := range normalizedPhones {
		phoneValues = append(phoneValues, phone.Phone)
	}
	excludeID := customerID
	return normalized, model.CustomerUniqueCheckInput{
		ExcludeCustomerID: &excludeID,
		Name:              normalized.Name,
		Weixin:            normalized.Weixin,
		Phones:            phoneValues,
	}, nil
}

func normalizePhones(phones []model.CustomerPhoneInput) ([]model.CustomerPhoneInput, error) {
	normalized := make([]model.CustomerPhoneInput, 0, len(phones))
	seen := make(map[string]struct{})
	primaryIndex := -1
	for _, item := range phones {
		phone := normalizeCustomerPhone(item.Phone)
		if phone == "" {
			continue
		}
		if !isValidCustomerPhone(phone) {
			return nil, ErrInvalidPhoneFormat
		}
		if _, exists := seen[phone]; exists {
			return nil, ErrCustomerPhoneExists
		}
		seen[phone] = struct{}{}
		normalized = append(normalized, model.CustomerPhoneInput{
			Phone:      phone,
			PhoneLabel: strings.TrimSpace(item.PhoneLabel),
			IsPrimary:  item.IsPrimary,
		})
		if item.IsPrimary && primaryIndex < 0 {
			primaryIndex = len(normalized) - 1
		}
	}

	if len(normalized) == 0 {
		return nil, ErrInvalidPhoneFormat
	}
	if primaryIndex < 0 {
		normalized[0].IsPrimary = true
	} else {
		for idx := range normalized {
			normalized[idx].IsPrimary = idx == primaryIndex
		}
	}

	return normalized, nil
}

func normalizeCustomerPhone(phone string) string {
	normalized := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, strings.TrimSpace(phone))
	if len(normalized) == 13 && strings.HasPrefix(normalized, "86") && customerMobilePhoneRegex.MatchString(normalized[2:]) {
		return normalized[2:]
	}
	return normalized
}

func isValidCustomerPhone(phone string) bool {
	return customerMobilePhoneRegex.MatchString(phone) || customerLandlinePhoneRegex.MatchString(phone)
}

func (s *customerService) logActivity(ctx context.Context, userID int64, action, targetType string, targetID int64, targetName, content string) {
	if s.activityLogRepo == nil {
		return
	}
	_ = s.activityLogRepo.Create(ctx, model.ActivityLog{
		UserID:     userID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		TargetName: targetName,
		Content:    content,
	})
}
