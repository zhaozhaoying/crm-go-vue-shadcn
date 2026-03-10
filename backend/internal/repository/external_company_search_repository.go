package repository

import (
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrExternalCompanySearchTaskNotFound = errors.New("external company search task not found")
)

type ExternalCompanySearchRepository interface {
	CreateTask(ctx context.Context, task *model.ExternalCompanySearchTask) (*model.ExternalCompanySearchTask, error)
	ListTasks(ctx context.Context, filter model.ExternalCompanySearchTaskListFilter) (model.ExternalCompanySearchTaskListResult, error)
	GetTaskByID(ctx context.Context, id int64) (*model.ExternalCompanySearchTask, error)
	ClaimNextRunnableTask(ctx context.Context, workerToken string, staleBefore time.Time) (*model.ExternalCompanySearchTask, error)
	UpdateTaskProgress(ctx context.Context, task *model.ExternalCompanySearchTask) error
	MarkTaskCompleted(ctx context.Context, taskID int64, workerToken string, finishedAt time.Time) error
	MarkTaskFailed(ctx context.Context, taskID int64, workerToken, message string, finishedAt time.Time) error
	CancelTask(ctx context.Context, taskID int64) error
	UpsertCompany(ctx context.Context, company *model.ExternalCompany) (*model.ExternalCompany, bool, error)
	SaveSearchResult(ctx context.Context, result *model.ExternalCompanySearchResult) (bool, error)
	AppendEvent(ctx context.Context, taskID int64, eventType, message, payload string) (*model.ExternalCompanySearchEvent, error)
	ListTaskEvents(ctx context.Context, taskID, afterSeq int64, limit int) ([]model.ExternalCompanySearchEvent, error)
	ListTaskResults(ctx context.Context, filter model.ExternalCompanySearchResultListFilter) (model.ExternalCompanySearchResultListResult, error)
}

type gormExternalCompanySearchRepository struct {
	db *gorm.DB
}

type externalCompanySearchResultRow struct {
	ID                int64        `gorm:"column:id"`
	TaskID            int64        `gorm:"column:task_id"`
	CompanyID         int64        `gorm:"column:company_id"`
	Platform          int          `gorm:"column:platform"`
	Keyword           string       `gorm:"column:keyword"`
	RegionKeyword     string       `gorm:"column:region_keyword"`
	PageNo            int          `gorm:"column:page_no"`
	RankNo            int          `gorm:"column:rank_no"`
	IsNewCompany      bool         `gorm:"column:is_new_company"`
	ResultPayload     string       `gorm:"column:result_payload"`
	CreatedAt         time.Time    `gorm:"column:created_at"`
	UpdatedAt         time.Time    `gorm:"column:updated_at"`
	CompanyNo         string       `gorm:"column:company_no"`
	PlatformCompanyID string       `gorm:"column:platform_company_id"`
	DedupeKey         string       `gorm:"column:dedupe_key"`
	CompanyName       string       `gorm:"column:company_name"`
	CompanyNameEn     string       `gorm:"column:company_name_en"`
	CompanyURL        string       `gorm:"column:company_url"`
	CompanyLogo       string       `gorm:"column:company_logo"`
	CompanyImages     string       `gorm:"column:company_images"`
	CompanyDesc       string       `gorm:"column:company_desc"`
	Country           string       `gorm:"column:country"`
	Province          string       `gorm:"column:province"`
	City              string       `gorm:"column:city"`
	Address           string       `gorm:"column:address"`
	MainProducts      string       `gorm:"column:main_products"`
	BusinessType      string       `gorm:"column:business_type"`
	EmployeeCount     string       `gorm:"column:employee_count"`
	EstablishedYear   string       `gorm:"column:established_year"`
	AnnualRevenue     string       `gorm:"column:annual_revenue"`
	Certification     string       `gorm:"column:certification"`
	Contact           string       `gorm:"column:contact"`
	Phone             string       `gorm:"column:phone"`
	Email             string       `gorm:"column:email"`
	DataVersion       int          `gorm:"column:data_version"`
	InterestStatus    int          `gorm:"column:interest_status"`
	IsDeleted         bool         `gorm:"column:is_deleted"`
	RawPayload        string       `gorm:"column:raw_payload"`
	FirstSeenAt       sql.NullTime `gorm:"column:first_seen_at"`
	LastSeenAt        sql.NullTime `gorm:"column:last_seen_at"`
}

func NewGormExternalCompanySearchRepository(db *gorm.DB) ExternalCompanySearchRepository {
	return &gormExternalCompanySearchRepository{db: db}
}

func (r *gormExternalCompanySearchRepository) CreateTask(ctx context.Context, task *model.ExternalCompanySearchTask) (*model.ExternalCompanySearchTask, error) {
	now := time.Now().UTC()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = now
	}
	if task.NextRunAt == nil {
		task.NextRunAt = &now
	}
	if err := r.db.WithContext(ctx).Table(task.TableName()).Create(task).Error; err != nil {
		return nil, err
	}
	return r.GetTaskByID(ctx, task.ID)
}

func (r *gormExternalCompanySearchRepository) ListTasks(ctx context.Context, filter model.ExternalCompanySearchTaskListFilter) (model.ExternalCompanySearchTaskListResult, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	query := r.db.WithContext(ctx).Table((model.ExternalCompanySearchTask{}).TableName())
	if filter.Platform > 0 {
		query = query.Where("platform = ?", filter.Platform)
	}
	if filter.Status > 0 {
		query = query.Where("status = ?", filter.Status)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("(keyword LIKE ? OR region_keyword LIKE ?)", like, like)
	}
	if filter.RestrictToCreator && filter.CreatedBy > 0 {
		query = query.Where("created_by = ?", filter.CreatedBy)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return model.ExternalCompanySearchTaskListResult{}, err
	}

	var items []model.ExternalCompanySearchTask
	if err := query.Order("created_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Scan(&items).Error; err != nil {
		return model.ExternalCompanySearchTaskListResult{}, err
	}

	return model.ExternalCompanySearchTaskListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (r *gormExternalCompanySearchRepository) GetTaskByID(ctx context.Context, id int64) (*model.ExternalCompanySearchTask, error) {
	var task model.ExternalCompanySearchTask
	if err := r.db.WithContext(ctx).Table(task.TableName()).Where("id = ?", id).Take(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExternalCompanySearchTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (r *gormExternalCompanySearchRepository) ClaimNextRunnableTask(ctx context.Context, workerToken string, staleBefore time.Time) (*model.ExternalCompanySearchTask, error) {
	for range 3 {
		var claimed model.ExternalCompanySearchTask
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			now := time.Now().UTC()
			candidateQuery := tx.Table((model.ExternalCompanySearchTask{}).TableName()).
				Where("next_run_at <= ?", now).
				Where("(status = ?) OR (status = ? AND locked_at IS NOT NULL AND locked_at <= ?)", model.ExternalCompanySearchTaskStatusPending, model.ExternalCompanySearchTaskStatusRunning, staleBefore).
				Order("priority ASC, id ASC").
				Limit(1)
			if err := candidateQuery.Take(&claimed).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				return err
			}

			updates := map[string]any{
				"status":            model.ExternalCompanySearchTaskStatusRunning,
				"worker_token":      workerToken,
				"locked_at":         now,
				"last_heartbeat_at": now,
				"updated_at":        now,
			}
			if claimed.StartedAt == nil {
				updates["started_at"] = now
			}

			result := tx.Table(claimed.TableName()).
				Where("id = ?", claimed.ID).
				Where("(status = ?) OR (status = ? AND locked_at IS NOT NULL AND locked_at <= ?)", model.ExternalCompanySearchTaskStatusPending, model.ExternalCompanySearchTaskStatusRunning, staleBefore).
				Updates(updates)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				claimed.ID = 0
				return nil
			}
			return tx.Table(claimed.TableName()).Where("id = ?", claimed.ID).Take(&claimed).Error
		})
		if err != nil {
			return nil, err
		}
		if claimed.ID > 0 {
			return &claimed, nil
		}
	}
	return nil, nil
}

func (r *gormExternalCompanySearchRepository) UpdateTaskProgress(ctx context.Context, task *model.ExternalCompanySearchTask) error {
	now := time.Now().UTC()
	updates := map[string]any{
		"status":            task.Status,
		"page_no":           task.PageNo,
		"progress_percent":  task.ProgressPercent,
		"fetched_count":     task.FetchedCount,
		"saved_count":       task.SavedCount,
		"duplicate_count":   task.DuplicateCount,
		"failed_count":      task.FailedCount,
		"resume_cursor":     task.ResumeCursor,
		"error_message":     task.ErrorMessage,
		"last_heartbeat_at": now,
		"updated_at":        now,
	}
	query := r.db.WithContext(ctx).Table(task.TableName()).Where("id = ?", task.ID)
	if strings.TrimSpace(task.WorkerToken) != "" {
		query = query.Where("worker_token = ?", task.WorkerToken)
	}
	result := query.Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrExternalCompanySearchTaskNotFound
	}
	return nil
}

func (r *gormExternalCompanySearchRepository) MarkTaskCompleted(ctx context.Context, taskID int64, workerToken string, finishedAt time.Time) error {
	updates := map[string]any{
		"status":            model.ExternalCompanySearchTaskStatusCompleted,
		"progress_percent":  100,
		"finished_at":       finishedAt,
		"locked_at":         nil,
		"worker_token":      "",
		"last_heartbeat_at": finishedAt,
		"updated_at":        finishedAt,
	}
	result := r.db.WithContext(ctx).Table((model.ExternalCompanySearchTask{}).TableName()).Where("id = ?", taskID)
	if strings.TrimSpace(workerToken) != "" {
		result = result.Where("worker_token = ?", workerToken)
	}
	tx := result.Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrExternalCompanySearchTaskNotFound
	}
	return nil
}

func (r *gormExternalCompanySearchRepository) MarkTaskFailed(ctx context.Context, taskID int64, workerToken, message string, finishedAt time.Time) error {
	updates := map[string]any{
		"status":            model.ExternalCompanySearchTaskStatusFailed,
		"error_message":     strings.TrimSpace(message),
		"finished_at":       finishedAt,
		"locked_at":         nil,
		"worker_token":      "",
		"last_heartbeat_at": finishedAt,
		"updated_at":        finishedAt,
	}
	query := r.db.WithContext(ctx).Table((model.ExternalCompanySearchTask{}).TableName()).Where("id = ?", taskID)
	if strings.TrimSpace(workerToken) != "" {
		query = query.Where("worker_token = ?", workerToken)
	}
	result := query.Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrExternalCompanySearchTaskNotFound
	}
	return nil
}

func (r *gormExternalCompanySearchRepository) CancelTask(ctx context.Context, taskID int64) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Table((model.ExternalCompanySearchTask{}).TableName()).
		Where("id = ?", taskID).
		Where("status IN ?", []int{model.ExternalCompanySearchTaskStatusPending, model.ExternalCompanySearchTaskStatusRunning}).
		Updates(map[string]any{
			"status":       model.ExternalCompanySearchTaskStatusCanceled,
			"finished_at":  now,
			"locked_at":    nil,
			"worker_token": "",
			"updated_at":   now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrExternalCompanySearchTaskNotFound
	}
	return nil
}

func (r *gormExternalCompanySearchRepository) UpsertCompany(ctx context.Context, company *model.ExternalCompany) (*model.ExternalCompany, bool, error) {
	now := time.Now().UTC()
	created := false
	resultCompany := &model.ExternalCompany{}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.ExternalCompany
		err := tx.Table(company.TableName()).Where("platform = ? AND dedupe_key = ?", company.Platform, company.DedupeKey).Take(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				created = true
				if company.CompanyNo == "" {
					return errors.New("company_no is required")
				}
				company.DataVersion = maxInt(company.DataVersion, 1)
				company.InterestStatus = maxInt(company.InterestStatus, 1)
				company.FirstSeenAt = &now
				company.LastSeenAt = &now
				company.CreateTime = now
				company.UpdateTime = now
				if err := tx.Table(company.TableName()).Create(company).Error; err != nil {
					return err
				}
				*resultCompany = *company
				return nil
			}
			return err
		}

		merged := mergeExternalCompany(existing, *company, now)
		if err := tx.Table(existing.TableName()).Where("id = ?", existing.ID).Updates(merged).Error; err != nil {
			return err
		}
		if err := tx.Table(existing.TableName()).Where("id = ?", existing.ID).Take(resultCompany).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return resultCompany, created, nil
}

func (r *gormExternalCompanySearchRepository) SaveSearchResult(ctx context.Context, result *model.ExternalCompanySearchResult) (bool, error) {
	created := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.ExternalCompanySearchResult
		err := tx.Table(result.TableName()).Where("task_id = ? AND company_id = ?", result.TaskID, result.CompanyID).Take(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				created = true
				now := time.Now().UTC()
				result.CreatedAt = now
				result.UpdatedAt = now
				return tx.Table(result.TableName()).Create(result).Error
			}
			return err
		}
		return tx.Table(existing.TableName()).Where("id = ?", existing.ID).Updates(map[string]any{
			"page_no":        result.PageNo,
			"rank_no":        result.RankNo,
			"is_new_company": result.IsNewCompany,
			"result_payload": chooseString(result.ResultPayload, existing.ResultPayload),
			"updated_at":     time.Now().UTC(),
		}).Error
	})
	return created, err
}

func (r *gormExternalCompanySearchRepository) AppendEvent(ctx context.Context, taskID int64, eventType, message, payload string) (*model.ExternalCompanySearchEvent, error) {
	event := &model.ExternalCompanySearchEvent{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxSeq sql.NullInt64
		if err := tx.Table((model.ExternalCompanySearchEvent{}).TableName()).Select("MAX(seq_no)").Where("task_id = ?", taskID).Scan(&maxSeq).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		*event = model.ExternalCompanySearchEvent{
			TaskID:    taskID,
			SeqNo:     maxSeq.Int64 + 1,
			EventType: strings.TrimSpace(eventType),
			Message:   strings.TrimSpace(message),
			Payload:   payload,
			CreatedAt: now,
		}
		return tx.Table(event.TableName()).Create(event).Error
	})
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (r *gormExternalCompanySearchRepository) ListTaskEvents(ctx context.Context, taskID, afterSeq int64, limit int) ([]model.ExternalCompanySearchEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	var items []model.ExternalCompanySearchEvent
	query := r.db.WithContext(ctx).Table((model.ExternalCompanySearchEvent{}).TableName()).Where("task_id = ?", taskID)
	if afterSeq > 0 {
		query = query.Where("seq_no > ?", afterSeq)
	}
	if err := query.Order("seq_no ASC").Limit(limit).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *gormExternalCompanySearchRepository) ListTaskResults(ctx context.Context, filter model.ExternalCompanySearchResultListFilter) (model.ExternalCompanySearchResultListResult, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	base := r.db.WithContext(ctx).
		Table("external_company_search_result AS r").
		Joins("JOIN external_company AS c ON c.id = r.company_id")
	if filter.TaskID > 0 {
		base = base.Where("r.task_id = ?", filter.TaskID)
	}
	if filter.Platform > 0 {
		base = base.Where("r.platform = ?", filter.Platform)
	}
	if filter.NewOnly {
		base = base.Where("r.is_new_company = ?", true)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + search + "%"
		base = base.Where(
			"(c.company_name LIKE ? OR c.company_name_en LIKE ? OR r.keyword LIKE ? OR r.region_keyword LIKE ? OR c.company_desc LIKE ? OR c.main_products LIKE ? OR c.business_type LIKE ? OR c.city LIKE ? OR c.province LIKE ? OR c.country LIKE ?)",
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
		)
	}
	if filter.RestrictToCreator && filter.CreatedBy > 0 {
		base = base.
			Joins("JOIN external_company_search_task AS t ON t.id = r.task_id").
			Where("t.created_by = ?", filter.CreatedBy)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return model.ExternalCompanySearchResultListResult{}, err
	}

	const selectColumns = `
		r.id,
		r.task_id,
		r.company_id,
		r.platform,
		r.keyword,
		r.region_keyword,
		r.page_no,
		r.rank_no,
		r.is_new_company,
		r.result_payload,
		r.created_at,
		r.updated_at,
		c.company_no,
		c.platform_company_id,
		c.dedupe_key,
		c.company_name,
		c.company_name_en,
		c.company_url,
		c.company_logo,
		c.company_images,
		c.company_desc,
		c.country,
		c.province,
		c.city,
		c.address,
		c.main_products,
		c.business_type,
		c.employee_count,
		c.established_year,
		c.annual_revenue,
		c.certification,
		c.contact,
		c.phone,
		c.email,
		c.data_version,
		c.interest_status,
		c.is_deleted,
		c.raw_payload,
		c.first_seen_at,
		c.last_seen_at`

	var rows []externalCompanySearchResultRow
	if err := base.Select(selectColumns).Order("r.id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Scan(&rows).Error; err != nil {
		return model.ExternalCompanySearchResultListResult{}, err
	}

	items := make([]model.ExternalCompanySearchResultItem, 0, len(rows))
	for _, row := range rows {
		item := model.ExternalCompanySearchResultItem{
			ID:                row.ID,
			TaskID:            row.TaskID,
			CompanyID:         row.CompanyID,
			Platform:          row.Platform,
			Keyword:           row.Keyword,
			RegionKeyword:     row.RegionKeyword,
			PageNo:            row.PageNo,
			RankNo:            row.RankNo,
			IsNewCompany:      row.IsNewCompany,
			ResultPayload:     row.ResultPayload,
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
			CompanyNo:         row.CompanyNo,
			PlatformCompanyID: row.PlatformCompanyID,
			DedupeKey:         row.DedupeKey,
			CompanyName:       row.CompanyName,
			CompanyNameEn:     row.CompanyNameEn,
			CompanyURL:        row.CompanyURL,
			CompanyLogo:       row.CompanyLogo,
			CompanyImages:     row.CompanyImages,
			CompanyDesc:       row.CompanyDesc,
			Country:           row.Country,
			Province:          row.Province,
			City:              row.City,
			Address:           row.Address,
			MainProducts:      row.MainProducts,
			BusinessType:      row.BusinessType,
			EmployeeCount:     row.EmployeeCount,
			EstablishedYear:   row.EstablishedYear,
			AnnualRevenue:     row.AnnualRevenue,
			Certification:     row.Certification,
			Contact:           row.Contact,
			Phone:             row.Phone,
			Email:             row.Email,
			DataVersion:       row.DataVersion,
			InterestStatus:    row.InterestStatus,
			IsDeleted:         row.IsDeleted,
			RawPayload:        row.RawPayload,
		}
		if row.FirstSeenAt.Valid {
			value := row.FirstSeenAt.Time
			item.FirstSeenAt = &value
		}
		if row.LastSeenAt.Valid {
			value := row.LastSeenAt.Time
			item.LastSeenAt = &value
		}
		items = append(items, item)
	}

	return model.ExternalCompanySearchResultListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func mergeExternalCompany(existing, incoming model.ExternalCompany, now time.Time) map[string]any {
	return map[string]any{
		"platform_company_id": chooseString(incoming.PlatformCompanyID, existing.PlatformCompanyID),
		"company_name":        chooseString(incoming.CompanyName, existing.CompanyName),
		"company_name_en":     chooseString(incoming.CompanyNameEn, existing.CompanyNameEn),
		"company_url":         chooseString(incoming.CompanyURL, existing.CompanyURL),
		"company_logo":        chooseString(incoming.CompanyLogo, existing.CompanyLogo),
		"company_images":      chooseString(incoming.CompanyImages, existing.CompanyImages),
		"company_desc":        chooseString(incoming.CompanyDesc, existing.CompanyDesc),
		"country":             chooseString(incoming.Country, existing.Country),
		"province":            chooseString(incoming.Province, existing.Province),
		"city":                chooseString(incoming.City, existing.City),
		"address":             chooseString(incoming.Address, existing.Address),
		"main_products":       chooseString(incoming.MainProducts, existing.MainProducts),
		"business_type":       chooseString(incoming.BusinessType, existing.BusinessType),
		"employee_count":      chooseString(incoming.EmployeeCount, existing.EmployeeCount),
		"established_year":    chooseString(incoming.EstablishedYear, existing.EstablishedYear),
		"annual_revenue":      chooseString(incoming.AnnualRevenue, existing.AnnualRevenue),
		"certification":       chooseString(incoming.Certification, existing.Certification),
		"contact":             chooseString(incoming.Contact, existing.Contact),
		"phone":               chooseString(incoming.Phone, existing.Phone),
		"email":               chooseString(incoming.Email, existing.Email),
		"raw_payload":         chooseString(incoming.RawPayload, existing.RawPayload),
		"data_version":        maxInt(existing.DataVersion+1, 1),
		"last_seen_at":        now,
		"update_time":         now,
	}
}

func chooseString(incoming, existing string) string {
	if strings.TrimSpace(incoming) != "" {
		return strings.TrimSpace(incoming)
	}
	return existing
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
