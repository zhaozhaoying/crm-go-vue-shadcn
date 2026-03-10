package service

import (
	"backend/internal/external/companysearch"
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var errExternalCompanySearchTargetReached = errors.New("external company search target reached")

type ExternalCompanySearchHub struct {
	mu          sync.RWMutex
	subscribers map[int64]map[chan model.ExternalCompanySearchEvent]struct{}
	bufferSize  int
}

func NewExternalCompanySearchHub(bufferSize int) *ExternalCompanySearchHub {
	if bufferSize <= 0 {
		bufferSize = 64
	}
	return &ExternalCompanySearchHub{
		subscribers: make(map[int64]map[chan model.ExternalCompanySearchEvent]struct{}),
		bufferSize:  bufferSize,
	}
}

func (h *ExternalCompanySearchHub) Subscribe(taskID int64) (<-chan model.ExternalCompanySearchEvent, func()) {
	ch := make(chan model.ExternalCompanySearchEvent, h.bufferSize)
	h.mu.Lock()
	if _, ok := h.subscribers[taskID]; !ok {
		h.subscribers[taskID] = make(map[chan model.ExternalCompanySearchEvent]struct{})
	}
	h.subscribers[taskID][ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		subscribers := h.subscribers[taskID]
		if subscribers == nil {
			return
		}
		if _, ok := subscribers[ch]; !ok {
			return
		}
		delete(subscribers, ch)
		close(ch)
		if len(subscribers) == 0 {
			delete(h.subscribers, taskID)
		}
	}
	return ch, unsubscribe
}

func (h *ExternalCompanySearchHub) Publish(event model.ExternalCompanySearchEvent) {
	h.mu.RLock()
	subscribers := h.subscribers[event.TaskID]
	channels := make([]chan model.ExternalCompanySearchEvent, 0, len(subscribers))
	for ch := range subscribers {
		channels = append(channels, ch)
	}
	h.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- event:
		default:
		}
	}
}

type ExternalCompanySearchRuntime struct {
	repo         repository.ExternalCompanySearchRepository
	hub          *ExternalCompanySearchHub
	providers    map[int]companysearch.Provider
	workerCount  int
	pollInterval time.Duration
	wakeCh       chan struct{}
}

type externalCompanySearchProgressPayload struct {
	TaskID          int64 `json:"taskId"`
	Status          int   `json:"status"`
	PageNo          int   `json:"pageNo"`
	ProgressPercent int   `json:"progressPercent"`
	FetchedCount    int   `json:"fetchedCount"`
	SavedCount      int   `json:"savedCount"`
	DuplicateCount  int   `json:"duplicateCount"`
	FailedCount     int   `json:"failedCount"`
}

func NewExternalCompanySearchRuntime(
	repo repository.ExternalCompanySearchRepository,
	hub *ExternalCompanySearchHub,
	workerCount int,
	pollInterval time.Duration,
	providers ...companysearch.Provider,
) *ExternalCompanySearchRuntime {
	providerMap := make(map[int]companysearch.Provider, len(providers))
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		providerMap[provider.Platform()] = provider
	}
	if workerCount <= 0 {
		workerCount = 1
	}
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	return &ExternalCompanySearchRuntime{
		repo:         repo,
		hub:          hub,
		providers:    providerMap,
		workerCount:  workerCount,
		pollInterval: pollInterval,
		wakeCh:       make(chan struct{}, workerCount),
	}
}

func (r *ExternalCompanySearchRuntime) Publish(event model.ExternalCompanySearchEvent) {
	if r == nil || r.hub == nil {
		return
	}
	r.hub.Publish(event)
}

func (r *ExternalCompanySearchRuntime) Wake() {
	if r == nil {
		return
	}
	select {
	case r.wakeCh <- struct{}{}:
	default:
	}
}

func (r *ExternalCompanySearchRuntime) Start(ctx context.Context) {
	for index := 0; index < r.workerCount; index++ {
		go r.workerLoop(ctx, index+1)
	}
}

func (r *ExternalCompanySearchRuntime) workerLoop(ctx context.Context, workerIndex int) {
	workerToken := fmt.Sprintf("external-company-search-worker-%d", workerIndex)
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		processed, err := r.processNext(ctx, workerToken)
		if err != nil {
			log.Printf("external company search worker=%s process next failed: %v", workerToken, err)
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-r.wakeCh:
		}
	}
}

func (r *ExternalCompanySearchRuntime) processNext(ctx context.Context, workerToken string) (bool, error) {
	staleBefore := time.Now().UTC().Add(-2 * time.Minute)
	task, err := r.repo.ClaimNextRunnableTask(ctx, workerToken, staleBefore)
	if err != nil {
		return false, err
	}
	if task == nil {
		return false, nil
	}
	if _, err := r.appendEvent(ctx, task.ID, model.ExternalCompanySearchEventTaskStarted, "task started", map[string]any{
		"taskId": task.ID,
		"status": task.Status,
	}); err != nil {
		log.Printf("append start event failed for task=%d: %v", task.ID, err)
	}
	if err := r.executeTask(ctx, task); err != nil {
		return true, err
	}
	return true, nil
}

func (r *ExternalCompanySearchRuntime) executeTask(ctx context.Context, task *model.ExternalCompanySearchTask) error {
	provider, ok := r.providers[task.Platform]
	if !ok {
		message := fmt.Sprintf("provider for platform %d is not configured", task.Platform)
		return r.failTask(ctx, task, message)
	}

	request := companysearch.SearchRequest{
		TaskID:        task.ID,
		Keyword:       task.Keyword,
		RegionKeyword: task.RegionKeyword,
		PageLimit:     task.PageLimit,
		TargetCount:   task.TargetCount,
		SearchOptions: task.SearchOptions,
	}

	err := provider.Search(ctx, request, func(page companysearch.SearchPage) error {
		if r.isTaskCanceled(ctx, task) {
			return context.Canceled
		}
		return r.handleSearchPage(ctx, task, page)
	})
	if err != nil {
		if errors.Is(err, errExternalCompanySearchTargetReached) {
			return r.completeTask(ctx, task)
		}
		if errors.Is(err, context.Canceled) && r.isTaskCanceled(ctx, task) {
			return nil
		}
		return r.failTask(ctx, task, err.Error())
	}
	return r.completeTask(ctx, task)
}

func (r *ExternalCompanySearchRuntime) handleSearchPage(ctx context.Context, task *model.ExternalCompanySearchTask, page companysearch.SearchPage) error {
	task.PageNo = page.PageNo
	task.ResumeCursor = page.ResumeCursor
	for _, item := range page.Items {
		task.FetchedCount++
		if err := r.handleFetchedCompany(ctx, task, item); err != nil {
			task.FailedCount++
			log.Printf("process fetched company failed, task=%d page=%d rank=%d: %v", task.ID, item.PageNo, item.RankNo, err)
		}
		if task.TargetCount > 0 && task.SavedCount >= task.TargetCount {
			break
		}
	}
	if page.EstimatedTotalPages > 0 {
		task.ProgressPercent = minInt(99, (page.PageNo*100)/page.EstimatedTotalPages)
	}
	if task.TargetCount > 0 && task.SavedCount >= task.TargetCount {
		task.ProgressPercent = minInt(99, task.ProgressPercent)
	}
	if err := r.repo.UpdateTaskProgress(ctx, task); err != nil {
		return err
	}
	_, _ = r.appendEvent(ctx, task.ID, model.ExternalCompanySearchEventTaskProgress, "task progress updated", externalCompanySearchProgressPayload{
		TaskID:          task.ID,
		Status:          task.Status,
		PageNo:          task.PageNo,
		ProgressPercent: task.ProgressPercent,
		FetchedCount:    task.FetchedCount,
		SavedCount:      task.SavedCount,
		DuplicateCount:  task.DuplicateCount,
		FailedCount:     task.FailedCount,
	})
	if task.TargetCount > 0 && task.SavedCount >= task.TargetCount {
		return errExternalCompanySearchTargetReached
	}
	return nil
}

func (r *ExternalCompanySearchRuntime) handleFetchedCompany(ctx context.Context, task *model.ExternalCompanySearchTask, item companysearch.FetchedCompany) error {
	company := &model.ExternalCompany{
		CompanyNo:         companysearch.NewCompanyNo(),
		Platform:          task.Platform,
		PlatformCompanyID: item.PlatformCompanyID,
		DedupeKey:         item.DedupeKey,
		CompanyName:       item.CompanyName,
		CompanyNameEn:     item.CompanyNameEn,
		CompanyURL:        item.CompanyURL,
		CompanyLogo:       item.CompanyLogo,
		CompanyImages:     item.CompanyImages,
		CompanyDesc:       item.CompanyDesc,
		Country:           item.Country,
		Province:          item.Province,
		City:              item.City,
		Address:           item.Address,
		MainProducts:      item.MainProducts,
		BusinessType:      item.BusinessType,
		EmployeeCount:     item.EmployeeCount,
		EstablishedYear:   item.EstablishedYear,
		AnnualRevenue:     item.AnnualRevenue,
		Certification:     item.Certification,
		Contact:           item.Contact,
		Phone:             item.Phone,
		Email:             item.Email,
		DataVersion:       1,
		InterestStatus:    1,
		RawPayload:        item.RawPayload,
	}
	storedCompany, isNewCompany, err := r.repo.UpsertCompany(ctx, company)
	if err != nil {
		return err
	}
	if !isNewCompany {
		task.DuplicateCount++
	}

	result := &model.ExternalCompanySearchResult{
		TaskID:        task.ID,
		CompanyID:     storedCompany.ID,
		Platform:      task.Platform,
		Keyword:       task.Keyword,
		RegionKeyword: task.RegionKeyword,
		PageNo:        item.PageNo,
		RankNo:        item.RankNo,
		IsNewCompany:  isNewCompany,
		ResultPayload: item.ResultPayload,
	}
	resultCreated, err := r.repo.SaveSearchResult(ctx, result)
	if err != nil {
		return err
	}
	if resultCreated {
		task.SavedCount++
	}
	if resultCreated {
		_, _ = r.appendEvent(ctx, task.ID, model.ExternalCompanySearchEventResultSaved, "result saved", map[string]any{
			"taskId":         task.ID,
			"companyId":      storedCompany.ID,
			"companyName":    storedCompany.CompanyName,
			"platform":       storedCompany.Platform,
			"pageNo":         item.PageNo,
			"rankNo":         item.RankNo,
			"isNewCompany":   isNewCompany,
			"duplicateCount": task.DuplicateCount,
		})
	}
	return nil
}

func (r *ExternalCompanySearchRuntime) completeTask(ctx context.Context, task *model.ExternalCompanySearchTask) error {
	now := time.Now().UTC()
	if err := r.repo.MarkTaskCompleted(ctx, task.ID, task.WorkerToken, now); err != nil {
		return err
	}
	_, _ = r.appendEvent(ctx, task.ID, model.ExternalCompanySearchEventTaskCompleted, "task completed", externalCompanySearchProgressPayload{
		TaskID:          task.ID,
		Status:          model.ExternalCompanySearchTaskStatusCompleted,
		PageNo:          task.PageNo,
		ProgressPercent: 100,
		FetchedCount:    task.FetchedCount,
		SavedCount:      task.SavedCount,
		DuplicateCount:  task.DuplicateCount,
		FailedCount:     task.FailedCount,
	})
	return nil
}

func (r *ExternalCompanySearchRuntime) failTask(ctx context.Context, task *model.ExternalCompanySearchTask, message string) error {
	now := time.Now().UTC()
	if err := r.repo.MarkTaskFailed(ctx, task.ID, task.WorkerToken, message, now); err != nil {
		return err
	}
	_, _ = r.appendEvent(ctx, task.ID, model.ExternalCompanySearchEventTaskFailed, message, map[string]any{
		"taskId":         task.ID,
		"status":         model.ExternalCompanySearchTaskStatusFailed,
		"pageNo":         task.PageNo,
		"fetchedCount":   task.FetchedCount,
		"savedCount":     task.SavedCount,
		"duplicateCount": task.DuplicateCount,
		"failedCount":    task.FailedCount,
		"errorMessage":   message,
	})
	return nil
}

func (r *ExternalCompanySearchRuntime) appendEvent(ctx context.Context, taskID int64, eventType, message string, payload any) (*model.ExternalCompanySearchEvent, error) {
	event, err := r.repo.AppendEvent(ctx, taskID, eventType, message, marshalExternalCompanySearchPayload(payload))
	if err != nil {
		return nil, err
	}
	r.Publish(*event)
	return event, nil
}

func (r *ExternalCompanySearchRuntime) isTaskCanceled(ctx context.Context, task *model.ExternalCompanySearchTask) bool {
	if task == nil {
		return false
	}
	freshTask, err := r.repo.GetTaskByID(ctx, task.ID)
	if err != nil {
		return false
	}
	task.Status = freshTask.Status
	return freshTask.Status == model.ExternalCompanySearchTaskStatusCanceled
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
