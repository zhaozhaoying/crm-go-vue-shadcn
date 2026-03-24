package service

import (
	"backend/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrCallRecordingNotFound = errors.New("call recording not found")

type callRecordingRepository interface {
	List(ctx context.Context, filter model.CallRecordingListFilter) (model.CallRecordingListResult, error)
	FindByID(ctx context.Context, id string, showAll bool, viewerHanghangCRMMobile string) (*model.CallRecording, error)
	GetLatestStartTime(ctx context.Context) (int64, error)
	UpsertBatch(ctx context.Context, items []model.CallRecordingUpsertInput) ([]model.CallRecording, error)
}

type CallRecordingService struct {
	repo callRecordingRepository
}

func NewCallRecordingService(repo callRecordingRepository) *CallRecordingService {
	return &CallRecordingService{repo: repo}
}

func (s *CallRecordingService) List(
	ctx context.Context,
	filter model.CallRecordingListFilter,
) (model.CallRecordingListResult, error) {
	result, err := s.repo.List(ctx, filter)
	if err != nil {
		return result, err
	}
	if result.Items == nil {
		result.Items = []model.CallRecording{}
	}
	return result, nil
}

func (s *CallRecordingService) GetByID(
	ctx context.Context,
	id string,
	showAll bool,
	viewerHanghangCRMMobile string,
) (*model.CallRecording, error) {
	item, err := s.repo.FindByID(ctx, id, showAll, viewerHanghangCRMMobile)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCallRecordingNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *CallRecordingService) UpsertBatch(
	ctx context.Context,
	items []model.CallRecordingUpsertInput,
) ([]model.CallRecording, error) {
	return s.repo.UpsertBatch(ctx, items)
}

func (s *CallRecordingService) GetLatestStartTime(ctx context.Context) (int64, error) {
	return s.repo.GetLatestStartTime(ctx)
}
