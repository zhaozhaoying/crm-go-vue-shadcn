package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"strings"
	"time"
)

type DashboardService interface {
	GetOverview(ctx context.Context, actorUserID int64, actorRole string) (model.DashboardOverview, error)
}

type dashboardService struct {
	repo repository.DashboardRepository
}

func NewDashboardService(repo repository.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetOverview(ctx context.Context, actorUserID int64, actorRole string) (model.DashboardOverview, error) {
	return s.repo.GetOverview(ctx, time.Now(), actorUserID, isDashboardGlobalRole(actorRole))
}

func isDashboardGlobalRole(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}
