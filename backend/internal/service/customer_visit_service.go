package service

import (
	"backend/internal/model"
	"context"
	"errors"
	"strings"
)

var ErrCustomerVisitAlreadyCheckedInToday = errors.New("customer visit already checked in today")

type customerVisitRepository interface {
	Create(input model.CustomerVisitCreateInput) (int64, error)
	ExistsSameDayVisitByUserCompanyIP(operatorUserID int64, customerName, checkInIP, visitDate string) (bool, error)
	List(filter model.CustomerVisitListFilter) (model.CustomerVisitListResult, error)
}

type CustomerVisitService struct {
	repo     customerVisitRepository
	resolver CustomerVisitLocationResolver
}

func NewCustomerVisitService(repo customerVisitRepository, resolver CustomerVisitLocationResolver) *CustomerVisitService {
	return &CustomerVisitService{
		repo:     repo,
		resolver: resolver,
	}
}

// Create 创建上门拜访记录
func (s *CustomerVisitService) Create(ctx context.Context, input model.CustomerVisitCreateInput) (int64, error) {
	input.CustomerName = strings.TrimSpace(input.CustomerName)
	input.Inviter = strings.TrimSpace(input.Inviter)
	input.CheckInIP = strings.TrimSpace(input.CheckInIP)
	input.DetailAddress = strings.TrimSpace(input.DetailAddress)
	input.VisitPurpose = strings.TrimSpace(input.VisitPurpose)
	input.Remark = strings.TrimSpace(input.Remark)
	input.Province = strings.TrimSpace(input.Province)
	input.City = strings.TrimSpace(input.City)
	input.Area = strings.TrimSpace(input.Area)

	if input.CustomerName != "" && input.CheckInIP != "" && input.VisitDate != "" {
		exists, err := s.repo.ExistsSameDayVisitByUserCompanyIP(
			input.OperatorUserID,
			input.CustomerName,
			input.CheckInIP,
			input.VisitDate,
		)
		if err != nil {
			return 0, err
		}
		if exists {
			return 0, ErrCustomerVisitAlreadyCheckedInToday
		}
	}

	hasResolvedAddress := input.Province != "" || input.City != "" || input.Area != "" || input.DetailAddress != ""
	if !hasResolvedAddress && s.resolver != nil && input.CheckInLat != 0 && input.CheckInLng != 0 {
		location, err := s.resolver.Resolve(ctx, input.CheckInLat, input.CheckInLng)
		if err != nil {
			if input.Province == "" && input.City == "" && input.Area == "" && input.DetailAddress == "" {
				return 0, err
			}
		} else {
			if location.Province != "" {
				input.Province = location.Province
			}
			if location.City != "" {
				input.City = location.City
			}
			if location.Area != "" {
				input.Area = location.Area
			}
			if strings.TrimSpace(location.DetailAddress) != "" {
				input.DetailAddress = strings.TrimSpace(location.DetailAddress)
			}
		}
	}

	return s.repo.Create(input)
}

// List 获取上门拜访记录列表
func (s *CustomerVisitService) List(filter model.CustomerVisitListFilter) (model.CustomerVisitListResult, error) {
	result, err := s.repo.List(filter)
	if err != nil {
		return result, err
	}
	if result.Items == nil {
		result.Items = []model.CustomerVisit{}
	}
	return result, nil
}
