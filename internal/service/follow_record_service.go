package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
)

type FollowRecordService struct {
	repo            *repository.FollowRecordRepository
	activityLogRepo *repository.ActivityLogRepository
}

func NewFollowRecordService(repo *repository.FollowRecordRepository, activityLogRepo ...*repository.ActivityLogRepository) *FollowRecordService {
	svc := &FollowRecordService{repo: repo}
	if len(activityLogRepo) > 0 {
		svc.activityLogRepo = activityLogRepo[0]
	}
	return svc
}

func ensureOperationCustomer(items []model.OperationFollowRecord) {
	for i := range items {
		if items[i].Customer == nil {
			items[i].Customer = &model.OperationFollowRecordCustomer{ID: items[i].CustomerID}
			continue
		}
		if items[i].Customer.ID == 0 {
			items[i].Customer.ID = items[i].CustomerID
		}
	}
}

func ensureSalesCustomer(items []model.SalesFollowRecord) {
	for i := range items {
		if items[i].Customer == nil {
			items[i].Customer = &model.OperationFollowRecordCustomer{ID: items[i].CustomerID}
			continue
		}
		if items[i].Customer.ID == 0 {
			items[i].Customer.ID = items[i].CustomerID
		}
	}
}

func (s *FollowRecordService) CreateOperationFollowRecord(input model.FollowRecordCreateInput) (int64, error) {
	id, err := s.repo.CreateOperationFollowRecord(input)
	if err != nil {
		return 0, err
	}
	s.logActivity(input.OperatorUserID, model.ActionOperationFollow, model.TargetTypeCustomer, input.CustomerID, "", input.Content)
	return id, nil
}

func (s *FollowRecordService) CreateSalesFollowRecord(input model.FollowRecordCreateInput) (int64, error) {
	id, err := s.repo.CreateSalesFollowRecord(input)
	if err != nil {
		return 0, err
	}
	s.logActivity(input.OperatorUserID, model.ActionSalesFollow, model.TargetTypeCustomer, input.CustomerID, "", input.Content)
	return id, nil
}

// ListOperationFollowRecords 获取运营跟进记录列表（按客户ID）
func (s *FollowRecordService) ListOperationFollowRecords(filter model.FollowRecordListFilter) (model.OperationFollowRecordListResult, error) {
	result, err := s.repo.ListOperationFollowRecords(filter)
	if err != nil {
		return result, err
	}
	ensureOperationCustomer(result.Items)
	return result, nil
}

// ListAllOperationFollowRecords 获取所有运营跟进记录列表
func (s *FollowRecordService) ListAllOperationFollowRecords(page, pageSize int) (model.OperationFollowRecordListResult, error) {
	result, err := s.repo.ListAllOperationFollowRecords(page, pageSize)
	if err != nil {
		return result, err
	}
	ensureOperationCustomer(result.Items)
	return result, nil
}

// ListSalesFollowRecords 获取销售跟进记录列表（按客户ID）
func (s *FollowRecordService) ListSalesFollowRecords(filter model.FollowRecordListFilter) (model.SalesFollowRecordListResult, error) {
	result, err := s.repo.ListSalesFollowRecords(filter)
	if err != nil {
		return result, err
	}
	ensureSalesCustomer(result.Items)
	return result, nil
}

// ListAllSalesFollowRecords 获取所有销售跟进记录列表
func (s *FollowRecordService) ListAllSalesFollowRecords(page, pageSize int) (model.SalesFollowRecordListResult, error) {
	result, err := s.repo.ListAllSalesFollowRecords(page, pageSize)
	if err != nil {
		return result, err
	}
	ensureSalesCustomer(result.Items)
	return result, nil
}

// ListFollowMethods 获取所有跟进方式
func (s *FollowRecordService) ListFollowMethods() ([]model.FollowMethod, error) {
	return s.repo.ListFollowMethods()
}

// CreateFollowMethod 创建跟进方式
func (s *FollowRecordService) CreateFollowMethod(req model.FollowMethodRequest) (int64, error) {
	return s.repo.CreateFollowMethod(req)
}

// UpdateFollowMethod 更新跟进方式
func (s *FollowRecordService) UpdateFollowMethod(id int, req model.FollowMethodRequest) error {
	return s.repo.UpdateFollowMethod(id, req)
}

func (s *FollowRecordService) DeleteFollowMethod(id int) error {
	return s.repo.DeleteFollowMethod(id)
}

func (s *FollowRecordService) logActivity(userID int64, action, targetType string, targetID int64, targetName, content string) {
	if s.activityLogRepo == nil {
		return
	}
	_ = s.activityLogRepo.Create(context.Background(), model.ActivityLog{
		UserID:     userID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		TargetName: targetName,
		Content:    content,
	})
}
