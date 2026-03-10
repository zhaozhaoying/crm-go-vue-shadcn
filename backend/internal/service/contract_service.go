package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrContractNotFound                    = errors.New("contract not found")
	ErrContractForbidden                   = errors.New("contract access forbidden")
	ErrContractNumberExists                = errors.New("contract number already exists")
	ErrContractContractNumberRequired      = errors.New("contract number is required")
	ErrContractNameRequired                = errors.New("contract name is required")
	ErrContractInvalidUser                 = errors.New("invalid user")
	ErrContractInvalidCustomer             = errors.New("invalid customer")
	ErrContractInvalidServiceUser          = errors.New("invalid service user")
	ErrContractInvalidCooperationType      = errors.New("invalid cooperation type")
	ErrContractInvalidPaymentStatus        = errors.New("invalid payment status")
	ErrContractInvalidAuditStatus          = errors.New("invalid audit status")
	ErrContractInvalidExpiryHandlingStatus = errors.New("invalid expiry handling status")
	ErrContractInvalidAmount               = errors.New("invalid amount")
	ErrContractPaymentExceedsContract      = errors.New("payment amount exceeds contract amount")
	ErrContractInvalidDateRange            = errors.New("end date cannot be earlier than start date")
	ErrContractAuditStatusForbidden        = errors.New("only admin or finance manager can update audit status")
	ErrContractNumberForbidden             = errors.New("only admin can update contract number")
	ErrContractAuditedReadonly             = errors.New("audited contract is readonly")
)

type ContractService interface {
	ListContracts(ctx context.Context, filter model.ContractListFilter) (model.ContractListResult, error)
	GetContractByID(ctx context.Context, id, actorUserID int64, actorRole string) (*model.Contract, error)
	IsContractNumberAvailable(ctx context.Context, contractNumber string, excludeID int64) (bool, error)
	CreateContract(ctx context.Context, input model.ContractCreateInput) (*model.Contract, error)
	UpdateContract(ctx context.Context, id int64, input model.ContractUpdateInput, actorUserID int64, actorRole string) (*model.Contract, error)
	AuditContract(ctx context.Context, id int64, input model.ContractUpdateInput, actorUserID int64, actorRole string) (*model.Contract, error)
	DeleteContract(ctx context.Context, id, actorUserID int64, actorRole string) error
}

type contractService struct {
	repo            repository.ContractRepository
	settingRepo     *repository.SystemSettingRepository
	activityLogRepo *repository.ActivityLogRepository
	defaultPrefix   string
}

func NewContractService(repo repository.ContractRepository, settingRepo *repository.SystemSettingRepository, activityLogRepo *repository.ActivityLogRepository) ContractService {
	return &contractService{repo: repo, settingRepo: settingRepo, activityLogRepo: activityLogRepo, defaultPrefix: "zzy_"}
}

func (s *contractService) ListContracts(ctx context.Context, filter model.ContractListFilter) (model.ContractListResult, error) {
	normalized := filter
	normalized.Keyword = strings.TrimSpace(filter.Keyword)
	normalized.PaymentStatus = strings.TrimSpace(filter.PaymentStatus)
	normalized.CooperationType = strings.TrimSpace(filter.CooperationType)
	normalized.AuditStatus = strings.TrimSpace(filter.AuditStatus)
	normalized.ExpiryHandlingStatus = strings.TrimSpace(filter.ExpiryHandlingStatus)

	scoped, err := s.applyListScopeByRole(ctx, normalized)
	if err != nil {
		return model.ContractListResult{}, err
	}
	normalized = scoped

	return s.repo.List(ctx, normalized)
}

func (s *contractService) GetContractByID(ctx context.Context, id, actorUserID int64, actorRole string) (*model.Contract, error) {
	return s.getAccessibleContract(ctx, id, actorUserID, actorRole)
}

func (s *contractService) getAccessibleContract(ctx context.Context, id, actorUserID int64, actorRole string) (*model.Contract, error) {
	contract, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrContractNotFound) {
			return nil, ErrContractNotFound
		}
		return nil, err
	}
	scope, err := s.resolveAccessScope(ctx, actorUserID, actorRole)
	if err != nil {
		return nil, err
	}
	if !canAccessContractByScope(contract, scope) {
		return nil, ErrContractForbidden
	}
	return contract, nil
}

func (s *contractService) IsContractNumberAvailable(ctx context.Context, contractNumber string, excludeID int64) (bool, error) {
	number := strings.TrimSpace(contractNumber)
	if number == "" {
		return false, ErrContractContractNumberRequired
	}
	exists, err := s.repo.ExistsContractNumber(ctx, number, excludeID)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func (s *contractService) CreateContract(ctx context.Context, input model.ContractCreateInput) (*model.Contract, error) {
	prefix := s.getContractNumberPrefix()
	input.ContractNumber = composeContractNumber(prefix, input.ContractNumber)
	input.AuditStatus = model.ContractAuditStatusPending
	input.AuditComment = ""
	input.AuditedBy = nil
	input.AuditedAt = nil

	normalized, err := s.normalizeCreateInput(ctx, input)
	if err != nil {
		return nil, err
	}
	contract, err := s.repo.Create(ctx, normalized)
	if err != nil {
		if errors.Is(err, repository.ErrContractNumberExists) {
			return nil, ErrContractNumberExists
		}
		return nil, err
	}
	s.logActivity(ctx, input.UserID, model.ActionCreateContract, model.TargetTypeContract, contract.ID, contract.ContractName, "")
	return contract, nil
}

func (s *contractService) UpdateContract(ctx context.Context, id int64, input model.ContractUpdateInput, actorUserID int64, actorRole string) (*model.Contract, error) {
	existing, err := s.getAccessibleContract(ctx, id, actorUserID, actorRole)
	if err != nil {
		return nil, err
	}
	if existing.AuditStatus != model.ContractAuditStatusPending && !isOperationRole(actorRole) {
		return nil, ErrContractAuditedReadonly
	}

	input = applyRoleScopedUpdateInput(existing, input, actorRole)

	prefix := s.getContractNumberPrefix()
	existingSuffix := extractSuffix(prefix, existing.ContractNumber)
	incomingSuffix := extractSuffix(prefix, input.ContractNumber)
	if incomingSuffix == "" {
		incomingSuffix = strings.TrimSpace(input.ContractNumber)
	}
	if incomingSuffix != existingSuffix && !isAdminRole(actorRole) {
		return nil, ErrContractNumberForbidden
	}
	input.ContractNumber = composeContractNumber(prefix, incomingSuffix)

	normalized, err := s.normalizeUpdateInput(ctx, input)
	if err != nil {
		return nil, err
	}
	if normalized.AuditStatus != existing.AuditStatus {
		return nil, ErrContractAuditStatusForbidden
	}
	normalized.AuditStatus = existing.AuditStatus
	normalized.AuditComment = existing.AuditComment
	normalized.AuditedBy = existing.AuditedBy
	normalized.AuditedAt = existing.AuditedAt

	contract, err := s.repo.Update(ctx, id, normalized)
	if err != nil {
		if errors.Is(err, repository.ErrContractNotFound) {
			return nil, ErrContractNotFound
		}
		if errors.Is(err, repository.ErrContractNumberExists) {
			return nil, ErrContractNumberExists
		}
		return nil, err
	}
	return contract, nil
}

func (s *contractService) AuditContract(ctx context.Context, id int64, input model.ContractUpdateInput, actorUserID int64, actorRole string) (*model.Contract, error) {
	if !canAuditContractRole(actorRole) {
		return nil, ErrContractAuditStatusForbidden
	}
	existing, err := s.getAccessibleContract(ctx, id, actorUserID, actorRole)
	if err != nil {
		return nil, err
	}
	if existing.AuditStatus != model.ContractAuditStatusPending {
		return nil, ErrContractAuditedReadonly
	}

	prefix := s.getContractNumberPrefix()
	existingSuffix := extractSuffix(prefix, existing.ContractNumber)
	incomingSuffix := extractSuffix(prefix, input.ContractNumber)
	if incomingSuffix == "" {
		incomingSuffix = strings.TrimSpace(input.ContractNumber)
	}
	if incomingSuffix != existingSuffix && !isAdminRole(actorRole) {
		return nil, ErrContractNumberForbidden
	}
	input.ContractNumber = composeContractNumber(prefix, incomingSuffix)

	normalized, err := s.normalizeUpdateInput(ctx, input)
	if err != nil {
		return nil, err
	}
	if normalized.AuditStatus != model.ContractAuditStatusSuccess && normalized.AuditStatus != model.ContractAuditStatusFailed {
		return nil, ErrContractInvalidAuditStatus
	}
	auditedAt := time.Now().UTC()
	normalized.AuditComment = strings.TrimSpace(input.AuditComment)
	normalized.AuditedBy = &actorUserID
	normalized.AuditedAt = &auditedAt

	contract, err := s.repo.Update(ctx, id, normalized)
	if err != nil {
		if errors.Is(err, repository.ErrContractNotFound) {
			return nil, ErrContractNotFound
		}
		if errors.Is(err, repository.ErrContractNumberExists) {
			return nil, ErrContractNumberExists
		}
		return nil, err
	}
	content := existing.AuditStatus + " → " + normalized.AuditStatus
	if normalized.AuditComment != "" {
		content += " | " + normalized.AuditComment
	}
	s.logActivity(ctx, actorUserID, model.ActionAuditContract, model.TargetTypeContract, contract.ID, contract.ContractName, content)
	return contract, nil
}

func (s *contractService) DeleteContract(ctx context.Context, id, actorUserID int64, actorRole string) error {
	if _, err := s.getAccessibleContract(ctx, id, actorUserID, actorRole); err != nil {
		return err
	}
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrContractNotFound) {
			return ErrContractNotFound
		}
		return err
	}
	return nil
}

func (s *contractService) normalizeCreateInput(ctx context.Context, input model.ContractCreateInput) (model.ContractCreateInput, error) {
	normalized := model.ContractCreateInput{
		ContractImage:        strings.TrimSpace(input.ContractImage),
		PaymentImage:         strings.TrimSpace(input.PaymentImage),
		PaymentStatus:        strings.TrimSpace(input.PaymentStatus),
		Remark:               strings.TrimSpace(input.Remark),
		UserID:               input.UserID,
		CustomerID:           input.CustomerID,
		CooperationType:      strings.TrimSpace(input.CooperationType),
		ContractNumber:       strings.TrimSpace(input.ContractNumber),
		ContractName:         strings.TrimSpace(input.ContractName),
		ContractAmount:       input.ContractAmount,
		PaymentAmount:        input.PaymentAmount,
		CooperationYears:     input.CooperationYears,
		NodeCount:            input.NodeCount,
		ServiceUserID:        input.ServiceUserID,
		WebsiteName:          strings.TrimSpace(input.WebsiteName),
		WebsiteURL:           strings.TrimSpace(input.WebsiteURL),
		WebsiteUsername:      strings.TrimSpace(input.WebsiteUsername),
		IsOnline:             input.IsOnline,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		AuditStatus:          strings.TrimSpace(input.AuditStatus),
		AuditComment:         strings.TrimSpace(input.AuditComment),
		AuditedBy:            input.AuditedBy,
		AuditedAt:            input.AuditedAt,
		ExpiryHandlingStatus: strings.TrimSpace(input.ExpiryHandlingStatus),
	}
	applyContractTimeline(normalized.IsOnline, normalized.CooperationYears, &normalized.StartDate, &normalized.EndDate)
	if err := s.validateAndFillDefaults(ctx, normalized.UserID, normalized.CustomerID, normalized.ServiceUserID, &normalized.CooperationType, &normalized.PaymentStatus, &normalized.AuditStatus, &normalized.ExpiryHandlingStatus, normalized.ContractNumber, normalized.ContractName, normalized.ContractAmount, normalized.PaymentAmount, normalized.StartDate, normalized.EndDate); err != nil {
		return model.ContractCreateInput{}, err
	}
	return normalized, nil
}

func (s *contractService) normalizeUpdateInput(ctx context.Context, input model.ContractUpdateInput) (model.ContractUpdateInput, error) {
	normalized := model.ContractUpdateInput{
		ContractImage:        strings.TrimSpace(input.ContractImage),
		PaymentImage:         strings.TrimSpace(input.PaymentImage),
		PaymentStatus:        strings.TrimSpace(input.PaymentStatus),
		Remark:               strings.TrimSpace(input.Remark),
		UserID:               input.UserID,
		CustomerID:           input.CustomerID,
		CooperationType:      strings.TrimSpace(input.CooperationType),
		ContractNumber:       strings.TrimSpace(input.ContractNumber),
		ContractName:         strings.TrimSpace(input.ContractName),
		ContractAmount:       input.ContractAmount,
		PaymentAmount:        input.PaymentAmount,
		CooperationYears:     input.CooperationYears,
		NodeCount:            input.NodeCount,
		ServiceUserID:        input.ServiceUserID,
		WebsiteName:          strings.TrimSpace(input.WebsiteName),
		WebsiteURL:           strings.TrimSpace(input.WebsiteURL),
		WebsiteUsername:      strings.TrimSpace(input.WebsiteUsername),
		IsOnline:             input.IsOnline,
		StartDate:            input.StartDate,
		EndDate:              input.EndDate,
		AuditStatus:          strings.TrimSpace(input.AuditStatus),
		AuditComment:         strings.TrimSpace(input.AuditComment),
		AuditedBy:            input.AuditedBy,
		AuditedAt:            input.AuditedAt,
		ExpiryHandlingStatus: strings.TrimSpace(input.ExpiryHandlingStatus),
	}
	applyContractTimeline(normalized.IsOnline, normalized.CooperationYears, &normalized.StartDate, &normalized.EndDate)
	if err := s.validateAndFillDefaults(ctx, normalized.UserID, normalized.CustomerID, normalized.ServiceUserID, &normalized.CooperationType, &normalized.PaymentStatus, &normalized.AuditStatus, &normalized.ExpiryHandlingStatus, normalized.ContractNumber, normalized.ContractName, normalized.ContractAmount, normalized.PaymentAmount, normalized.StartDate, normalized.EndDate); err != nil {
		return model.ContractUpdateInput{}, err
	}
	return normalized, nil
}

func applyRoleScopedUpdateInput(existing *model.Contract, input model.ContractUpdateInput, actorRole string) model.ContractUpdateInput {
	if existing == nil {
		return input
	}

	merged := model.ContractUpdateInput{
		ContractImage:        existing.ContractImage,
		PaymentImage:         existing.PaymentImage,
		PaymentStatus:        existing.PaymentStatus,
		Remark:               existing.Remark,
		UserID:               existing.UserID,
		CustomerID:           existing.CustomerID,
		CooperationType:      existing.CooperationType,
		ContractNumber:       existing.ContractNumber,
		ContractName:         existing.ContractName,
		ContractAmount:       existing.ContractAmount,
		PaymentAmount:        existing.PaymentAmount,
		CooperationYears:     existing.CooperationYears,
		NodeCount:            existing.NodeCount,
		ServiceUserID:        copyInt64Pointer(existing.ServiceUserID),
		WebsiteName:          existing.WebsiteName,
		WebsiteURL:           existing.WebsiteURL,
		WebsiteUsername:      existing.WebsiteUsername,
		IsOnline:             existing.IsOnline,
		StartDate:            copyInt64Pointer(existing.StartDateUnix),
		EndDate:              copyInt64Pointer(existing.EndDateUnix),
		AuditStatus:          existing.AuditStatus,
		AuditComment:         existing.AuditComment,
		AuditedBy:            copyInt64Pointer(existing.AuditedBy),
		AuditedAt:            copyTimePointer(existing.AuditedAt),
		ExpiryHandlingStatus: existing.ExpiryHandlingStatus,
	}

	switch {
	case isOperationRole(actorRole):
		merged.WebsiteName = input.WebsiteName
		merged.WebsiteURL = input.WebsiteURL
		merged.WebsiteUsername = input.WebsiteUsername
		merged.IsOnline = input.IsOnline
		return merged
	case isSalesRole(actorRole):
		merged.ContractImage = input.ContractImage
		merged.PaymentImage = input.PaymentImage
		merged.PaymentStatus = input.PaymentStatus
		merged.Remark = input.Remark
		merged.UserID = input.UserID
		merged.CustomerID = input.CustomerID
		merged.CooperationType = input.CooperationType
		merged.ContractNumber = input.ContractNumber
		merged.ContractName = input.ContractName
		merged.ContractAmount = input.ContractAmount
		merged.PaymentAmount = input.PaymentAmount
		merged.CooperationYears = input.CooperationYears
		merged.NodeCount = input.NodeCount
		merged.ServiceUserID = input.ServiceUserID
		merged.ExpiryHandlingStatus = input.ExpiryHandlingStatus
		return merged
	default:
		return input
	}
}

func applyContractTimeline(isOnline bool, cooperationYears int, startDate **int64, endDate **int64) {
	if startDate == nil || endDate == nil {
		return
	}

	if *startDate == nil && isOnline {
		now := time.Now().UTC().Unix()
		*startDate = &now
	}

	if *startDate == nil {
		*endDate = nil
		return
	}

	start := time.Unix(**startDate, 0).UTC()
	end := start.AddDate(cooperationYears, 0, 0).Unix()
	*endDate = &end
}

func copyInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func copyTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

type contractAccessScope struct {
	all                   bool
	userID                int64
	allowedUserIDs        []int64
	allowedServiceUserIDs []int64
	forceServiceUserID    *int64
}

func (s *contractService) resolveAccessScope(ctx context.Context, actorUserID int64, actorRole string) (contractAccessScope, error) {
	scope := contractAccessScope{}
	role := strings.TrimSpace(actorRole)
	if actorUserID <= 0 {
		return scope, nil
	}

	if isRole(role, "admin", "管理员", "finance_manager", "finance", "财务经理", "财务") {
		scope.all = true
		return scope, nil
	}

	if isRole(role, "sales_director", "销售总监") {
		managerIDs, err := s.repo.ListDirectSubordinateUserIDsByRoleNames(ctx, []int64{actorUserID}, []string{"sales_manager", "销售经理"})
		if err != nil {
			return contractAccessScope{}, err
		}
		staffIDs, err := s.repo.ListDirectSubordinateUserIDsByRoleNames(ctx, managerIDs, []string{"sales_staff", "销售员工", "销售"})
		if err != nil {
			return contractAccessScope{}, err
		}
		scope.allowedUserIDs = uniqueInt64(append([]int64{actorUserID}, append(managerIDs, staffIDs...)...))
		if len(scope.allowedUserIDs) == 0 {
			scope.allowedUserIDs = []int64{-1}
		}
		return scope, nil
	}

	if isRole(role, "ops_staff", "operation_staff", "运营员工", "运营") {
		force := actorUserID
		scope.forceServiceUserID = &force
		return scope, nil
	}

	if isRole(role, "ops_manager", "operation_manager", "运营经理") {
		ids, err := s.repo.ListUserIDsByRoleNames(ctx, []string{"ops_manager", "ops_staff", "运营经理", "运营员工", "运营"})
		if err != nil {
			return contractAccessScope{}, err
		}
		scope.allowedServiceUserIDs = uniqueInt64(ids)
		if len(scope.allowedServiceUserIDs) == 0 {
			scope.allowedServiceUserIDs = []int64{-1}
		}
		return scope, nil
	}

	scope.userID = actorUserID
	return scope, nil
}

func (s *contractService) validateAndFillDefaults(
	ctx context.Context,
	userID int64,
	customerID int64,
	serviceUserID *int64,
	cooperationType *string,
	paymentStatus *string,
	auditStatus *string,
	expiryHandlingStatus *string,
	contractNumber string,
	contractName string,
	contractAmount float64,
	paymentAmount float64,
	startDate *int64,
	endDate *int64,
) error {
	if contractNumber == "" {
		return ErrContractContractNumberRequired
	}
	if contractName == "" {
		return ErrContractNameRequired
	}
	if contractAmount < 0 || paymentAmount < 0 {
		return ErrContractInvalidAmount
	}
	if paymentAmount > contractAmount {
		return ErrContractPaymentExceedsContract
	}
	if startDate != nil && endDate != nil && *endDate < *startDate {
		return ErrContractInvalidDateRange
	}

	if *cooperationType == "" {
		*cooperationType = model.ContractCooperationTypeDomestic
	}
	if *paymentStatus == "" {
		*paymentStatus = model.ContractPaymentStatusPending
	}
	if *auditStatus == "" {
		*auditStatus = model.ContractAuditStatusPending
	}
	if *expiryHandlingStatus == "" {
		*expiryHandlingStatus = model.ContractExpiryHandlingStatusPending
	}

	if *cooperationType != model.ContractCooperationTypeDomestic && *cooperationType != model.ContractCooperationTypeForeign {
		return ErrContractInvalidCooperationType
	}
	if *paymentStatus != model.ContractPaymentStatusPending && *paymentStatus != model.ContractPaymentStatusPaid && *paymentStatus != model.ContractPaymentStatusPartial {
		return ErrContractInvalidPaymentStatus
	}
	if *auditStatus != model.ContractAuditStatusPending && *auditStatus != model.ContractAuditStatusSuccess && *auditStatus != model.ContractAuditStatusFailed {
		return ErrContractInvalidAuditStatus
	}
	if *expiryHandlingStatus != model.ContractExpiryHandlingStatusPending && *expiryHandlingStatus != model.ContractExpiryHandlingStatusRenewed && *expiryHandlingStatus != model.ContractExpiryHandlingStatusEnded {
		return ErrContractInvalidExpiryHandlingStatus
	}

	userExists, err := s.repo.ExistsUser(ctx, userID)
	if err != nil {
		return err
	}
	if !userExists {
		return ErrContractInvalidUser
	}
	customerExists, err := s.repo.ExistsCustomer(ctx, customerID)
	if err != nil {
		return err
	}
	if !customerExists {
		return ErrContractInvalidCustomer
	}
	if serviceUserID != nil {
		serviceExists, err := s.repo.ExistsUser(ctx, *serviceUserID)
		if err != nil {
			return err
		}
		if !serviceExists {
			return ErrContractInvalidServiceUser
		}
	}

	return nil
}

func isAdminRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "admin")
}

func isSalesRole(role string) bool {
	return isRole(role,
		"sales_director", "sales_manager", "sales_staff",
		"销售总监", "销售经理", "销售员工", "销售",
	)
}

func isOperationRole(role string) bool {
	return isRole(role,
		"ops_manager", "operation_manager", "ops_staff", "operation_staff",
		"运营经理", "运营员工", "运营",
	)
}

func canAuditContractRole(role string) bool {
	return isRole(role, "admin", "管理员", "finance_manager", "财务经理")
}

func isRole(role string, candidates ...string) bool {
	current := strings.TrimSpace(role)
	if current == "" {
		return false
	}
	for _, candidate := range candidates {
		if strings.EqualFold(current, strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func (s *contractService) applyListScopeByRole(ctx context.Context, filter model.ContractListFilter) (model.ContractListFilter, error) {
	scope, err := s.resolveAccessScope(ctx, filter.ActorUserID, filter.ActorRole)
	if err != nil {
		return model.ContractListFilter{}, err
	}
	if scope.all {
		return filter, nil
	}
	if len(scope.allowedUserIDs) > 0 {
		filter.AllowedUserIDs = scope.allowedUserIDs
		return filter, nil
	}
	if len(scope.allowedServiceUserIDs) > 0 {
		filter.AllowedServiceUserIDs = scope.allowedServiceUserIDs
		return filter, nil
	}
	if scope.forceServiceUserID != nil {
		filter.ForceServiceUserID = scope.forceServiceUserID
		return filter, nil
	}
	if scope.userID > 0 {
		filter.UserID = scope.userID
	}
	return filter, nil
}

func canAccessContractByScope(contract *model.Contract, scope contractAccessScope) bool {
	if contract == nil {
		return false
	}
	if scope.all {
		return true
	}
	if len(scope.allowedUserIDs) > 0 {
		return containsInt64(scope.allowedUserIDs, contract.UserID)
	}
	if len(scope.allowedServiceUserIDs) > 0 {
		return contract.ServiceUserID != nil && containsInt64(scope.allowedServiceUserIDs, *contract.ServiceUserID)
	}
	if scope.forceServiceUserID != nil {
		return contract.ServiceUserID != nil && *contract.ServiceUserID == *scope.forceServiceUserID
	}
	if scope.userID > 0 {
		return contract.UserID == scope.userID
	}
	return false
}

func containsInt64(values []int64, target int64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func uniqueInt64(ids []int64) []int64 {
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

func (s *contractService) getContractNumberPrefix() string {
	if s.settingRepo == nil {
		return s.defaultPrefix
	}
	setting, err := s.settingRepo.GetSetting("contract_number_prefix")
	if err != nil || setting == nil {
		return s.defaultPrefix
	}
	prefix := strings.TrimSpace(setting.Value)
	if prefix == "" {
		return s.defaultPrefix
	}
	return prefix
}

func composeContractNumber(prefix string, suffixOrRaw string) string {
	normalizedPrefix := strings.TrimSpace(prefix)
	if normalizedPrefix == "" {
		normalizedPrefix = "zzy_"
	}
	trimmed := strings.TrimSpace(suffixOrRaw)
	if strings.HasPrefix(trimmed, normalizedPrefix) {
		suffix := strings.TrimSpace(strings.TrimPrefix(trimmed, normalizedPrefix))
		return normalizedPrefix + suffix
	}
	return normalizedPrefix + trimmed
}

func extractSuffix(prefix, fullNumber string) string {
	normalizedPrefix := strings.TrimSpace(prefix)
	trimmed := strings.TrimSpace(fullNumber)
	if normalizedPrefix == "" {
		return trimmed
	}
	if strings.HasPrefix(trimmed, normalizedPrefix) {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, normalizedPrefix))
	}
	return trimmed
}

func (s *contractService) logActivity(ctx context.Context, userID int64, action, targetType string, targetID int64, targetName, content string) {
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
