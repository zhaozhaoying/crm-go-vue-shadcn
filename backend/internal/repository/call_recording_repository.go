package repository

import (
	"backend/internal/model"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CallRecordingRepository interface {
	List(ctx context.Context, filter model.CallRecordingListFilter) (model.CallRecordingListResult, error)
	FindByID(ctx context.Context, id string, showAll bool, viewerHanghangCRMMobile string) (*model.CallRecording, error)
	GetLatestStartTime(ctx context.Context) (int64, error)
	UpsertBatch(ctx context.Context, items []model.CallRecordingUpsertInput) ([]model.CallRecording, error)
}

type gormCallRecordingRepository struct {
	db *gorm.DB
}

func NewCallRecordingRepository(db *gorm.DB) CallRecordingRepository {
	return &gormCallRecordingRepository{db: db}
}

func (r *gormCallRecordingRepository) List(
	ctx context.Context,
	filter model.CallRecordingListFilter,
) (model.CallRecordingListResult, error) {
	result := model.CallRecordingListResult{}
	query := r.scopedQuery(r.db.WithContext(ctx).Model(&model.CallRecording{}), filter.ShowAll, filter.ViewerHanghangCRMMobile)
	if query == nil {
		return model.CallRecordingListResult{
			Items:    []model.CallRecording{},
			Total:    0,
			Page:     normalizePage(filter.Page),
			PageSize: normalizePageSize(filter.PageSize),
		}, nil
	}

	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where(
			`id LIKE ? OR real_name LIKE ? OR mobile LIKE ? OR phone LIKE ? OR tel_a LIKE ? OR tel_b LIKE ? OR enterprise_name LIKE ? OR dept_name LIKE ? OR interface_name LIKE ? OR line_name LIKE ?`,
			like, like, like, like, like, like, like, like, like, like,
		)
	}
	if filter.MinDuration > 0 {
		query = query.Where("duration >= ?", filter.MinDuration)
	}
	if filter.MaxDuration > 0 {
		query = query.Where("duration <= ?", filter.MaxDuration)
	}

	if err := query.Count(&result.Total).Error; err != nil {
		return result, err
	}

	page := normalizePage(filter.Page)
	pageSize := normalizePageSize(filter.PageSize)
	offset := (page - 1) * pageSize

	var items []model.CallRecording
	if err := query.
		Order("start_time DESC").
		Order("create_time DESC").
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return result, err
	}

	if items == nil {
		items = []model.CallRecording{}
	}
	result.Items = items
	result.Page = page
	result.PageSize = pageSize
	return result, nil
}

func (r *gormCallRecordingRepository) FindByID(
	ctx context.Context,
	id string,
	showAll bool,
	viewerHanghangCRMMobile string,
) (*model.CallRecording, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}

	query := r.scopedQuery(r.db.WithContext(ctx).Model(&model.CallRecording{}), showAll, viewerHanghangCRMMobile)
	if query == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var item model.CallRecording
	if err := query.Where("id = ?", id).Take(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormCallRecordingRepository) GetLatestStartTime(ctx context.Context) (int64, error) {
	var latest int64
	if err := r.db.WithContext(ctx).
		Model(&model.CallRecording{}).
		Where("start_time > 0").
		Select("COALESCE(MAX(start_time), 0)").
		Scan(&latest).Error; err != nil {
		return 0, err
	}
	return latest, nil
}

func (r *gormCallRecordingRepository) UpsertBatch(
	ctx context.Context,
	items []model.CallRecordingUpsertInput,
) ([]model.CallRecording, error) {
	items = dedupeCallRecordingUpsertInputs(items)
	if len(items) == 0 {
		return []model.CallRecording{}, nil
	}

	existingIDsByDedupeKey, err := r.findExistingIDsByDedupeKeys(ctx, items)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	result := make([]model.CallRecording, 0, len(items))
	for _, item := range items {
		dedupeKey := buildCallRecordingDedupeKey(item)
		rowID := item.ID
		if existingID, ok := existingIDsByDedupeKey[dedupeKey]; ok && strings.TrimSpace(existingID) != "" {
			rowID = existingID
		}
		row := model.CallRecording{
			ID:               rowID,
			AgentCode:        item.AgentCode,
			CallStatus:       item.CallStatus,
			CallStatusName:   item.CallStatusName,
			CallType:         item.CallType,
			CalleeAttr:       item.CalleeAttr,
			CallerAttr:       item.CallerAttr,
			CreateTime:       item.CreateTime,
			DeptName:         item.DeptName,
			Duration:         item.Duration,
			EndTime:          item.EndTime,
			EnterpriseName:   item.EnterpriseName,
			FinishStatus:     item.FinishStatus,
			FinishStatusName: item.FinishStatusName,
			Handle:           item.Handle,
			InterfaceID:      item.InterfaceID,
			InterfaceName:    item.InterfaceName,
			LineName:         item.LineName,
			Mobile:           item.Mobile,
			Mode:             item.Mode,
			MoveBatchCode:    item.MoveBatchCode,
			OctCustomerID:    item.OctCustomerID,
			Phone:            item.Phone,
			Postage:          item.Postage,
			PreRecordURL:     item.PreRecordURL,
			RealName:         item.RealName,
			StartTime:        item.StartTime,
			Status:           item.Status,
			TelA:             item.TelA,
			TelB:             item.TelB,
			TelX:             item.TelX,
			TenantCode:       item.TenantCode,
			UpdateTime:       item.UpdateTime,
			UserID:           item.UserID,
			WorkNum:          item.WorkNum,
			DedupeKey:        dedupeKey,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if err := r.db.WithContext(ctx).Table("call_recordings").Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"agent_code":         row.AgentCode,
				"call_status":        row.CallStatus,
				"call_status_name":   row.CallStatusName,
				"call_type":          row.CallType,
				"callee_attr":        row.CalleeAttr,
				"caller_attr":        row.CallerAttr,
				"create_time":        row.CreateTime,
				"dept_name":          row.DeptName,
				"duration":           row.Duration,
				"end_time":           row.EndTime,
				"enterprise_name":    row.EnterpriseName,
				"finish_status":      row.FinishStatus,
				"finish_status_name": row.FinishStatusName,
				"handle":             row.Handle,
				"interface_id":       row.InterfaceID,
				"interface_name":     row.InterfaceName,
				"line_name":          row.LineName,
				"mobile":             row.Mobile,
				"mode":               row.Mode,
				"move_batch_code":    row.MoveBatchCode,
				"oct_customer_id":    row.OctCustomerID,
				"phone":              row.Phone,
				"postage":            row.Postage,
				"pre_record_url":     row.PreRecordURL,
				"real_name":          row.RealName,
				"start_time":         row.StartTime,
				"status":             row.Status,
				"tel_a":              row.TelA,
				"tel_b":              row.TelB,
				"tel_x":              row.TelX,
				"tenant_code":        row.TenantCode,
				"update_time":        row.UpdateTime,
				"user_id":            row.UserID,
				"work_num":           row.WorkNum,
				"dedupe_key":         row.DedupeKey,
				"updated_at":         row.UpdatedAt,
			}),
		}).Create(&row).Error; err != nil {
			return nil, err
		}

		var saved model.CallRecording
		if err := r.db.WithContext(ctx).Table("call_recordings").Where("id = ?", row.ID).Take(&saved).Error; err != nil {
			return nil, err
		}
		result = append(result, saved)
	}

	return result, nil
}

func (r *gormCallRecordingRepository) findExistingIDsByDedupeKeys(
	ctx context.Context,
	items []model.CallRecordingUpsertInput,
) (map[string]string, error) {
	keys := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		key := buildCallRecordingDedupeKey(item)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return map[string]string{}, nil
	}

	type keyRow struct {
		ID        string `gorm:"column:id"`
		DedupeKey string `gorm:"column:dedupe_key"`
	}

	var rows []keyRow
	if err := r.db.WithContext(ctx).
		Table("call_recordings").
		Select("id, dedupe_key").
		Where("dedupe_key IN ?", keys).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.DedupeKey)
		if key == "" {
			continue
		}
		if _, exists := result[key]; exists {
			continue
		}
		result[key] = strings.TrimSpace(row.ID)
	}
	return result, nil
}

func (r *gormCallRecordingRepository) scopedQuery(
	query *gorm.DB,
	showAll bool,
	viewerHanghangCRMMobile string,
) *gorm.DB {
	if showAll {
		return query
	}
	viewerHanghangCRMMobile = strings.TrimSpace(viewerHanghangCRMMobile)
	if viewerHanghangCRMMobile == "" {
		return nil
	}
	return query.Where("(mobile = ? OR tel_a = ?)", viewerHanghangCRMMobile, viewerHanghangCRMMobile)
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 10
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

func dedupeCallRecordingUpsertInputs(items []model.CallRecordingUpsertInput) []model.CallRecordingUpsertInput {
	if len(items) == 0 {
		return []model.CallRecordingUpsertInput{}
	}

	order := make([]string, 0, len(items))
	merged := make(map[string]model.CallRecordingUpsertInput, len(items))
	for _, item := range items {
		item = normalizeCallRecordingUpsertInput(item)
		idKey := strings.TrimSpace(item.ID)
		if idKey == "" {
			continue
		}
		businessKey := buildCallRecordingDedupeKey(item)
		key := businessKey
		if key == "" {
			key = idKey
		}
		if _, exists := merged[key]; !exists {
			order = append(order, key)
		}
		merged[key] = item
	}

	result := make([]model.CallRecordingUpsertInput, 0, len(order))
	for _, key := range order {
		result = append(result, merged[key])
	}
	return result
}

func normalizeCallRecordingUpsertInput(item model.CallRecordingUpsertInput) model.CallRecordingUpsertInput {
	item.ID = strings.TrimSpace(item.ID)
	item.CallStatusName = strings.TrimSpace(item.CallStatusName)
	item.CalleeAttr = strings.TrimSpace(item.CalleeAttr)
	item.CallerAttr = strings.TrimSpace(item.CallerAttr)
	item.DeptName = strings.TrimSpace(item.DeptName)
	item.EnterpriseName = strings.TrimSpace(item.EnterpriseName)
	item.FinishStatusName = strings.TrimSpace(item.FinishStatusName)
	item.InterfaceID = strings.TrimSpace(item.InterfaceID)
	item.InterfaceName = strings.TrimSpace(item.InterfaceName)
	item.LineName = strings.TrimSpace(item.LineName)
	item.Mobile = strings.TrimSpace(item.Mobile)
	item.Phone = strings.TrimSpace(item.Phone)
	item.PreRecordURL = strings.TrimSpace(item.PreRecordURL)
	item.RealName = strings.TrimSpace(item.RealName)
	item.TelA = strings.TrimSpace(item.TelA)
	item.TelB = strings.TrimSpace(item.TelB)
	item.TelX = strings.TrimSpace(item.TelX)
	item.TenantCode = strings.TrimSpace(item.TenantCode)
	item.UserID = strings.TrimSpace(item.UserID)
	return item
}

func buildCallRecordingDedupeKey(item model.CallRecordingUpsertInput) string {
	return fmt.Sprintf(
		"%d|%s|%s|%s|%s|%d|%d",
		item.StartTime,
		strings.TrimSpace(item.Mobile),
		strings.TrimSpace(item.Phone),
		strings.TrimSpace(item.TelA),
		strings.TrimSpace(item.TelB),
		item.CallType,
		item.Duration,
	)
}
