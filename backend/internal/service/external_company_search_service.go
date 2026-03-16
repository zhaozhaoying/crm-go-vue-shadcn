package service

import (
	"backend/internal/errmsg"
	"backend/internal/external/companysearch"
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrExternalCompanySearchKeywordRequired     = errors.New("external company search keyword is required")
	ErrExternalCompanySearchPlatformRequired    = errors.New("external company search platform is required")
	ErrExternalCompanySearchPlatformUnsupported = errors.New("external company search platform is unsupported")
	ErrExternalCompanySearchTaskForbidden       = errors.New("external company search task forbidden")
)

type ExternalCompanySearchQueue interface {
	Wake()
}

type ExternalCompanySearchEventPublisher interface {
	Publish(event model.ExternalCompanySearchEvent)
}

type ExternalCompanySearchService interface {
	CreateTasks(ctx context.Context, input model.ExternalCompanySearchTaskCreateInput, actorRole string) ([]model.ExternalCompanySearchTask, error)
	ListTasks(ctx context.Context, filter model.ExternalCompanySearchTaskListFilter, actorRole string) (model.ExternalCompanySearchTaskListResult, error)
	GetTask(ctx context.Context, taskID, viewerID int64, actorRole string) (*model.ExternalCompanySearchTask, error)
	ListResults(ctx context.Context, viewerID int64, actorRole string, filter model.ExternalCompanySearchResultListFilter) (model.ExternalCompanySearchResultListResult, error)
	ListTaskResults(ctx context.Context, taskID, viewerID int64, actorRole string, filter model.ExternalCompanySearchResultListFilter) (model.ExternalCompanySearchResultListResult, error)
	ListTaskEvents(ctx context.Context, taskID, viewerID int64, actorRole string, afterSeq int64, limit int) (model.ExternalCompanySearchEventListResult, error)
	CancelTask(ctx context.Context, taskID, viewerID int64, actorRole string) error
}

type externalCompanySearchService struct {
	repo      repository.ExternalCompanySearchRepository
	queue     ExternalCompanySearchQueue
	publisher ExternalCompanySearchEventPublisher
}

func NewExternalCompanySearchService(
	repo repository.ExternalCompanySearchRepository,
	queue ExternalCompanySearchQueue,
	publisher ExternalCompanySearchEventPublisher,
) ExternalCompanySearchService {
	return &externalCompanySearchService{repo: repo, queue: queue, publisher: publisher}
}

func (s *externalCompanySearchService) CreateTasks(ctx context.Context, input model.ExternalCompanySearchTaskCreateInput, actorRole string) ([]model.ExternalCompanySearchTask, error) {
	keyword := strings.TrimSpace(input.Keyword)
	if keyword == "" {
		return nil, ErrExternalCompanySearchKeywordRequired
	}
	platforms := uniquePlatforms(input.Platforms)
	if len(platforms) == 0 {
		return nil, ErrExternalCompanySearchPlatformRequired
	}
	for _, platform := range platforms {
		if !isSupportedExternalCompanySearchPlatform(platform) {
			return nil, fmt.Errorf("%w: %d", ErrExternalCompanySearchPlatformUnsupported, platform)
		}
	}

	pageLimit := input.PageLimit
	if pageLimit < 0 {
		pageLimit = 0
	}
	priority := input.Priority
	if priority <= 0 {
		priority = 100
	}
	regionKeyword := strings.TrimSpace(input.RegionKeyword)
	if regionKeyword == "" {
		regionKeyword = keyword
	}
	keywordNormalized := companysearch.NormalizeKeyword(keyword)
	now := time.Now().UTC()

	createdTasks := make([]model.ExternalCompanySearchTask, 0, len(platforms))
	for _, platform := range platforms {
		task := &model.ExternalCompanySearchTask{
			TaskNo:            companysearch.NewTaskNo(),
			Platform:          platform,
			Keyword:           keyword,
			KeywordNormalized: keywordNormalized,
			RegionKeyword:     regionKeyword,
			Status:            model.ExternalCompanySearchTaskStatusPending,
			Priority:          priority,
			TargetCount:       input.TargetCount,
			PageLimit:         pageLimit,
			ProgressPercent:   0,
			MaxRetryCount:     0,
			SearchOptions:     strings.TrimSpace(input.SearchOptions),
			CreatedBy:         input.CreatedBy,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		saved, err := s.repo.CreateTask(ctx, task)
		if err != nil {
			return nil, err
		}
		createdTasks = append(createdTasks, *saved)
		_, _ = s.appendEvent(ctx, saved.ID, model.ExternalCompanySearchEventTaskCreated, "任务已创建", map[string]any{
			"task": saved,
		})
	}

	if s.queue != nil {
		s.queue.Wake()
	}
	_ = actorRole
	return createdTasks, nil
}

func (s *externalCompanySearchService) ListTasks(ctx context.Context, filter model.ExternalCompanySearchTaskListFilter, actorRole string) (model.ExternalCompanySearchTaskListResult, error) {
	filter = applyExternalCompanySearchTaskListScope(filter, actorRole)
	return s.repo.ListTasks(ctx, filter)
}

func (s *externalCompanySearchService) GetTask(ctx context.Context, taskID, viewerID int64, actorRole string) (*model.ExternalCompanySearchTask, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !canViewExternalCompanySearchTask(task, viewerID, actorRole) {
		return nil, ErrExternalCompanySearchTaskForbidden
	}
	return task, nil
}

func (s *externalCompanySearchService) ListResults(ctx context.Context, viewerID int64, actorRole string, filter model.ExternalCompanySearchResultListFilter) (model.ExternalCompanySearchResultListResult, error) {
	filter = applyExternalCompanySearchResultListScope(filter, viewerID, actorRole)
	return s.repo.ListTaskResults(ctx, filter)
}

func (s *externalCompanySearchService) ListTaskResults(ctx context.Context, taskID, viewerID int64, actorRole string, filter model.ExternalCompanySearchResultListFilter) (model.ExternalCompanySearchResultListResult, error) {
	if _, err := s.GetTask(ctx, taskID, viewerID, actorRole); err != nil {
		return model.ExternalCompanySearchResultListResult{}, err
	}
	filter.TaskID = taskID
	return s.repo.ListTaskResults(ctx, filter)
}

func (s *externalCompanySearchService) ListTaskEvents(ctx context.Context, taskID, viewerID int64, actorRole string, afterSeq int64, limit int) (model.ExternalCompanySearchEventListResult, error) {
	if _, err := s.GetTask(ctx, taskID, viewerID, actorRole); err != nil {
		return model.ExternalCompanySearchEventListResult{}, err
	}
	items, err := s.repo.ListTaskEvents(ctx, taskID, afterSeq, limit)
	if err != nil {
		return model.ExternalCompanySearchEventListResult{}, err
	}
	nextSeq := afterSeq
	if len(items) > 0 {
		nextSeq = items[len(items)-1].SeqNo
	}
	return model.ExternalCompanySearchEventListResult{Items: items, NextSeq: nextSeq}, nil
}

func (s *externalCompanySearchService) CancelTask(ctx context.Context, taskID, viewerID int64, actorRole string) error {
	task, err := s.GetTask(ctx, taskID, viewerID, actorRole)
	if err != nil {
		return err
	}
	if !canManageExternalCompanySearchTask(task, viewerID, actorRole) {
		return ErrExternalCompanySearchTaskForbidden
	}
	if task.Status == model.ExternalCompanySearchTaskStatusCompleted || task.Status == model.ExternalCompanySearchTaskStatusFailed || task.Status == model.ExternalCompanySearchTaskStatusCanceled {
		return nil
	}
	if err := s.repo.CancelTask(ctx, taskID); err != nil {
		return err
	}
	_, _ = s.appendEvent(ctx, taskID, model.ExternalCompanySearchEventTaskCanceled, "任务已取消", map[string]any{
		"taskId": taskID,
	})
	return nil
}

func (s *externalCompanySearchService) appendEvent(ctx context.Context, taskID int64, eventType, message string, payload any) (*model.ExternalCompanySearchEvent, error) {
	message = errmsg.Normalize(message)
	payloadJSON := marshalExternalCompanySearchPayload(payload)
	event, err := s.repo.AppendEvent(ctx, taskID, eventType, message, payloadJSON)
	if err != nil {
		return nil, err
	}
	if s.publisher != nil {
		s.publisher.Publish(*event)
	}
	return event, nil
}

func marshalExternalCompanySearchPayload(payload any) string {
	if payload == nil {
		return ""
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func uniquePlatforms(platforms []int) []int {
	seen := make(map[int]struct{}, len(platforms))
	result := make([]int, 0, len(platforms))
	for _, platform := range platforms {
		if platform <= 0 {
			continue
		}
		if _, exists := seen[platform]; exists {
			continue
		}
		seen[platform] = struct{}{}
		result = append(result, platform)
	}
	return result
}

func isSupportedExternalCompanySearchPlatform(platform int) bool {
	switch platform {
	case model.ExternalCompanyPlatformAlibaba, model.ExternalCompanyPlatformMadeInChina, model.ExternalCompanyPlatformGoogle:
		return true
	default:
		return false
	}
}

func canManageExternalCompanySearchTask(task *model.ExternalCompanySearchTask, viewerID int64, actorRole string) bool {
	if task == nil {
		return false
	}
	if isRole(actorRole, "admin", "管理员") {
		return true
	}
	return viewerID > 0 && task.CreatedBy == viewerID
}

func canViewExternalCompanySearchSharedData(actorRole string) bool {
	return isRole(actorRole,
		"admin", "管理员",
		"sales_director", "sales_manager", "sales_staff", "sales_inside", "sales_outside",
		"销售总监", "销售经理", "销售员工", "销售", "Inside销售", "Outside销售",
	)
}

func canViewExternalCompanySearchTask(task *model.ExternalCompanySearchTask, viewerID int64, actorRole string) bool {
	if task == nil {
		return false
	}
	if canViewExternalCompanySearchSharedData(actorRole) {
		return true
	}
	return viewerID > 0 && task.CreatedBy == viewerID
}

func applyExternalCompanySearchTaskListScope(filter model.ExternalCompanySearchTaskListFilter, actorRole string) model.ExternalCompanySearchTaskListFilter {
	if canViewExternalCompanySearchSharedData(actorRole) {
		filter.RestrictToCreator = false
		return filter
	}
	filter.RestrictToCreator = filter.CreatedBy > 0
	return filter
}

func applyExternalCompanySearchResultListScope(filter model.ExternalCompanySearchResultListFilter, viewerID int64, actorRole string) model.ExternalCompanySearchResultListFilter {
	if canViewExternalCompanySearchSharedData(actorRole) {
		filter.RestrictToCreator = false
		return filter
	}
	filter.CreatedBy = viewerID
	filter.RestrictToCreator = viewerID > 0
	return filter
}
