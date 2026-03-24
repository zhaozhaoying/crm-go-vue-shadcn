package service

import (
	"backend/internal/model"
	"context"
	"testing"
)

type contractRepoStub struct {
	existing        *model.Contract
	lastUpdateInput model.ContractUpdateInput
}

func (s *contractRepoStub) List(context.Context, model.ContractListFilter) (model.ContractListResult, error) {
	return model.ContractListResult{}, nil
}

func (s *contractRepoStub) GetByID(_ context.Context, _ int64) (*model.Contract, error) {
	if s.existing == nil {
		return nil, ErrContractNotFound
	}
	copied := *s.existing
	return &copied, nil
}

func (s *contractRepoStub) Create(context.Context, model.ContractCreateInput) (*model.Contract, error) {
	return nil, nil
}

func (s *contractRepoStub) Update(_ context.Context, _ int64, input model.ContractUpdateInput) (*model.Contract, error) {
	s.lastUpdateInput = input
	updated := *s.existing
	updated.UserID = input.UserID
	updated.Remark = input.Remark
	updated.AuditStatus = input.AuditStatus
	return &updated, nil
}

func (s *contractRepoStub) Delete(context.Context, int64) error {
	return nil
}

func (s *contractRepoStub) ExistsContractNumber(context.Context, string, int64) (bool, error) {
	return false, nil
}

func (s *contractRepoStub) ExistsUser(_ context.Context, id int64) (bool, error) {
	return id > 0, nil
}

func (s *contractRepoStub) ExistsCustomer(_ context.Context, id int64) (bool, error) {
	return id > 0, nil
}

func (s *contractRepoStub) ListUserIDsByRoleNames(context.Context, []string) ([]int64, error) {
	return []int64{}, nil
}

func (s *contractRepoStub) ListDirectSubordinateUserIDsByRoleNames(context.Context, []int64, []string) ([]int64, error) {
	return []int64{}, nil
}

func TestUpdateContractKeepsExistingSalesUserID(t *testing.T) {
	t.Parallel()

	repoStub := &contractRepoStub{
		existing: &model.Contract{
			ID:                   1,
			UserID:               7,
			CustomerID:           8,
			ContractNumber:       "zzy_001",
			ContractName:         "测试合同",
			ContractAmount:       100,
			PaymentAmount:        0,
			PaymentStatus:        model.ContractPaymentStatusPending,
			CooperationType:      model.ContractCooperationTypeDomestic,
			AuditStatus:          model.ContractAuditStatusPending,
			ExpiryHandlingStatus: model.ContractExpiryHandlingStatusPending,
		},
	}
	svc := &contractService{
		repo:          repoStub,
		defaultPrefix: "zzy_",
	}

	contract, err := svc.UpdateContract(context.Background(), 1, model.ContractUpdateInput{
		Remark:               "管理员编辑",
		CustomerID:           8,
		CooperationType:      model.ContractCooperationTypeDomestic,
		ContractNumber:       "001",
		ContractName:         "测试合同",
		ContractAmount:       100,
		PaymentAmount:        0,
		PaymentStatus:        model.ContractPaymentStatusPending,
		ExpiryHandlingStatus: model.ContractExpiryHandlingStatusPending,
	}, 99, "admin")
	if err != nil {
		t.Fatalf("UpdateContract returned error: %v", err)
	}

	if repoStub.lastUpdateInput.UserID != 7 {
		t.Fatalf("expected updated user_id to remain 7, got %d", repoStub.lastUpdateInput.UserID)
	}
	if contract.UserID != 7 {
		t.Fatalf("expected returned contract user_id to remain 7, got %d", contract.UserID)
	}
}

func TestAuditContractKeepsExistingSalesUserID(t *testing.T) {
	t.Parallel()

	repoStub := &contractRepoStub{
		existing: &model.Contract{
			ID:                   1,
			UserID:               7,
			CustomerID:           8,
			ContractNumber:       "zzy_001",
			ContractName:         "测试合同",
			ContractAmount:       100,
			PaymentAmount:        0,
			PaymentStatus:        model.ContractPaymentStatusPending,
			CooperationType:      model.ContractCooperationTypeDomestic,
			AuditStatus:          model.ContractAuditStatusPending,
			ExpiryHandlingStatus: model.ContractExpiryHandlingStatusPending,
		},
	}
	svc := &contractService{
		repo:          repoStub,
		defaultPrefix: "zzy_",
	}

	contract, err := svc.AuditContract(context.Background(), 1, model.ContractUpdateInput{
		Remark:               "财务审核",
		CustomerID:           8,
		CooperationType:      model.ContractCooperationTypeDomestic,
		ContractNumber:       "001",
		ContractName:         "测试合同",
		ContractAmount:       100,
		PaymentAmount:        0,
		PaymentStatus:        model.ContractPaymentStatusPending,
		AuditStatus:          model.ContractAuditStatusSuccess,
		ExpiryHandlingStatus: model.ContractExpiryHandlingStatusPending,
	}, 99, "admin")
	if err != nil {
		t.Fatalf("AuditContract returned error: %v", err)
	}

	if repoStub.lastUpdateInput.UserID != 7 {
		t.Fatalf("expected audited user_id to remain 7, got %d", repoStub.lastUpdateInput.UserID)
	}
	if contract.UserID != 7 {
		t.Fatalf("expected returned contract user_id to remain 7, got %d", contract.UserID)
	}
}
