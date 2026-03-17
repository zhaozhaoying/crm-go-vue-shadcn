package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
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
	return s.repo.GetOverview(ctx, time.Now(), actorUserID, actorRole)
}
