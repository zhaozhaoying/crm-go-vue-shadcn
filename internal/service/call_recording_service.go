package service

import (
	"backend/internal/model"
	"backend/internal/util"
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
	result.Items = maskCallRecordings(result.Items)
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
	if item == nil {
		return nil, nil
	}
	masked := maskCallRecording(*item)
	return &masked, nil
}

func (s *CallRecordingService) UpsertBatch(
	ctx context.Context,
	items []model.CallRecordingUpsertInput,
) ([]model.CallRecording, error) {
	saved, err := s.repo.UpsertBatch(ctx, items)
	if err != nil {
		return nil, err
	}
	return maskCallRecordings(saved), nil
}

func (s *CallRecordingService) GetLatestStartTime(ctx context.Context) (int64, error) {
	return s.repo.GetLatestStartTime(ctx)
}

func maskCallRecordings(items []model.CallRecording) []model.CallRecording {
	if len(items) == 0 {
		return items
	}

	masked := make([]model.CallRecording, len(items))
	for idx, item := range items {
		masked[idx] = maskCallRecording(item)
	}
	return masked
}

func maskCallRecording(item model.CallRecording) model.CallRecording {
	item.RealName = maskCallRecordingName(item.RealName)
	item.Mobile = util.MaskPhone(item.Mobile)
	item.Phone = util.MaskPhone(item.Phone)
	item.TelA = util.MaskPhone(item.TelA)
	item.TelB = util.MaskPhone(item.TelB)
	item.TelX = util.MaskPhone(item.TelX)
	return item
}

func maskCallRecordingName(name string) string {
	if name == "" {
		return name
	}
	return "**"
}
