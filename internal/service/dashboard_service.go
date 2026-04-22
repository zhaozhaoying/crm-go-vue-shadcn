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
	repo     repository.DashboardRepository
	location *time.Location
}

func NewDashboardService(repo repository.DashboardRepository, location ...*time.Location) DashboardService {
	service := &dashboardService{repo: repo, location: time.Local}
	if len(location) > 0 && location[0] != nil {
		service.location = location[0]
	}
	return service
}

func (s *dashboardService) GetOverview(ctx context.Context, actorUserID int64, actorRole string) (model.DashboardOverview, error) {
	return s.repo.GetOverview(ctx, time.Now().In(s.location), actorUserID, actorRole)
}
