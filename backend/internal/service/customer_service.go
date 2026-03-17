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
	ErrCustomerNameExists                   = errors.New("customer name already exists")
	ErrCustomerLegalExists                  = errors.New("customer legal name already exists")
	ErrCustomerWeixinExists                 = errors.New("customer weixin already exists")
	ErrCustomerPhoneExists                  = errors.New("customer phone already exists")
	ErrCustomerNameRequired                 = errors.New("customer name is required")
	ErrCustomerLimitExceeded                = errors.New("customer limit exceeded")
	ErrCustomerSameDepartmentClaimForbidden = errors.New("same department customer cannot be claimed")
	ErrCustomerNoOutsideSalesAvailable      = errors.New("no outside sales available")
	ErrPhoneNotFound                        = errors.New("phone not found")
	ErrPhoneAlreadyExists                   = errors.New("phone already exists for this customer")
	ErrInvalidPhoneFormat                   = errors.New("invalid phone format")
)

var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

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
	CreateCustomer(ctx context.Context, input model.CustomerCreateInput) (*model.Customer, error)
	UpdateCustomer(ctx context.Context, customerID int64, input model.CustomerUpdateInput) (*model.Customer, error)
	CheckUnique(ctx context.Context, input model.CustomerUniqueCheckInput) (model.CustomerUniqueCheckResult, error)
	ClaimCustomer(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error)
	ReleaseCustomer(ctx context.Context, customerID, operatorUserID int64) (*model.Customer, error)
	TransferCustomer(ctx context.Context, input model.CustomerTransferInput) (*model.Customer, error)

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
}

func (e *CustomerClaimFreezeError) Error() string {
	return "customer claim is frozen"
}

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
	// Inside-sales staff should also see customers they created and assigned to outside-sales.
	if isInsideSalesRole(role) {
		filter.IncludeCreatorScope = true
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
	normalized, err = s.applyCreateOwnerAssignment(ctx, normalized)
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
	if result.LegalNameExists {
		return nil, ErrCustomerLegalExists
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

func (s *customerService) applyCreateOwnerAssignment(ctx context.Context, input model.CustomerCreateInput) (model.CustomerCreateInput, error) {
	if strings.TrimSpace(input.Status) != model.CustomerStatusOwned || input.OperatorUserID <= 0 {
		return input, nil
	}

	operatorRole, err := s.repo.GetUserRoleName(ctx, input.OperatorUserID)
	if err != nil {
		return model.CustomerCreateInput{}, err
	}

	switch {
	case isInsideSalesRole(operatorRole):
		ownerUserID, err := pickBalancedSalesOwnerUserID(ctx, s.repo, input.OperatorUserID)
		if err != nil {
			return model.CustomerCreateInput{}, err
		}
		input.OwnerUserID = &ownerUserID
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
	if result.LegalNameExists {
		return nil, ErrCustomerLegalExists
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
		Weixin:            strings.TrimSpace(input.Weixin),
		Phones:            make([]string, 0, len(input.Phones)),
	}
	for _, phone := range input.Phones {
		trimmed := strings.TrimSpace(phone)
		if trimmed == "" {
			continue
		}
		normalized.Phones = append(normalized.Phones, trimmed)
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
		sameDepartment, err := s.isSameDepartment(ctx, operatorUserID, *customer.DropUserID)
		if err != nil {
			return nil, err
		}
		if sameDepartment {
			return nil, ErrCustomerSameDepartmentClaimForbidden
		}
	}

	if err := s.validateCustomerLimitByOwner(ctx, operatorUserID); err != nil {
		return nil, err
	}

	customer, err = s.repo.Claim(ctx, customerID, operatorUserID)
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

func (s *customerService) AddPhone(ctx context.Context, phone *model.CustomerPhone) error {
	// Validate phone format (Chinese mobile number)
	if !phoneRegex.MatchString(phone.Phone) {
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
	// Validate phone format
	if !phoneRegex.MatchString(phone.Phone) {
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

	phoneValues := make([]string, 0, len(normalizedPhones))
	for _, phone := range normalizedPhones {
		phoneValues = append(phoneValues, phone.Phone)
	}

	return normalized, model.CustomerUniqueCheckInput{
		Name:      normalized.Name,
		LegalName: normalized.LegalName,
		Weixin:    normalized.Weixin,
		Phones:    phoneValues,
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

	phoneValues := make([]string, 0, len(normalizedPhones))
	for _, phone := range normalizedPhones {
		phoneValues = append(phoneValues, phone.Phone)
	}
	excludeID := customerID
	return normalized, model.CustomerUniqueCheckInput{
		ExcludeCustomerID: &excludeID,
		Name:              normalized.Name,
		LegalName:         normalized.LegalName,
		Weixin:            normalized.Weixin,
		Phones:            phoneValues,
	}, nil
}

func normalizePhones(phones []model.CustomerPhoneInput) ([]model.CustomerPhoneInput, error) {
	normalized := make([]model.CustomerPhoneInput, 0, len(phones))
	seen := make(map[string]struct{})
	primaryIndex := -1
	for _, item := range phones {
		phone := strings.TrimSpace(item.Phone)
		if phone == "" {
			continue
		}
		if !phoneRegex.MatchString(phone) {
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
