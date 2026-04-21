package repository

import (
	"backend/internal/model"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TelemarketingRecordingRepository interface {
	List(ctx context.Context, filter model.TelemarketingRecordingListFilter) (model.TelemarketingRecordingListResult, error)
	FindByID(ctx context.Context, id string, showAll bool, viewerMihuaWorkNumber string) (*model.TelemarketingRecording, error)
	UpsertBatch(ctx context.Context, items []model.TelemarketingRecordingUpsertInput) ([]model.TelemarketingRecording, error)
	ListEnabledTelemarketingUsersByWorkNumbers(ctx context.Context, workNumbers []string) (map[string]model.TelemarketingRecordingMatchedUser, error)
}

type gormTelemarketingRecordingRepository struct {
	db *gorm.DB
}

const telemarketingRecordingIDMaxLen = 64

func NewTelemarketingRecordingRepository(db *gorm.DB) TelemarketingRecordingRepository {
	return &gormTelemarketingRecordingRepository{db: db}
}

func (r *gormTelemarketingRecordingRepository) List(
	ctx context.Context,
	filter model.TelemarketingRecordingListFilter,
) (model.TelemarketingRecordingListResult, error) {
	result := model.TelemarketingRecordingListResult{}
	query := r.scopedQuery(
		r.db.WithContext(ctx).Model(&model.TelemarketingRecording{}),
		filter.ShowAll,
		filter.ViewerMihuaWorkNo,
	)
	if query == nil {
		return model.TelemarketingRecordingListResult{
			Items:    []model.TelemarketingRecording{},
			Total:    0,
			Page:     normalizePage(filter.Page),
			PageSize: normalizePageSize(filter.PageSize),
		}, nil
	}

	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where(
			`cc_number LIKE ? OR outline_number LIKE ? OR service_seat_name LIKE ? OR service_seat_worknumber LIKE ? OR service_group_name LIKE ? OR service_number LIKE ? OR matched_user_name LIKE ? OR attribution LIKE ?`,
			like, like, like, like, like, like, like, like,
		)
	}
	if filter.MinDuration > 0 {
		query = query.Where("duration_second >= ?", filter.MinDuration)
	}
	if filter.MaxDuration > 0 {
		query = query.Where("duration_second <= ?", filter.MaxDuration)
	}

	startUnix, endUnixExclusive := telemarketingRecordingUnixRange(filter.StartDate, filter.EndDate)
	if startUnix > 0 {
		query = query.Where("initiate_time >= ?", startUnix)
	}
	if endUnixExclusive > 0 {
		query = query.Where("initiate_time < ?", endUnixExclusive)
	}

	if err := query.Count(&result.Total).Error; err != nil {
		return result, err
	}

	page := normalizePage(filter.Page)
	pageSize := normalizePageSize(filter.PageSize)
	offset := (page - 1) * pageSize

	var items []model.TelemarketingRecording
	if err := query.
		Order("initiate_time DESC").
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return result, err
	}
	if items == nil {
		items = []model.TelemarketingRecording{}
	}

	result.Items = items
	result.Page = page
	result.PageSize = pageSize
	return result, nil
}

func (r *gormTelemarketingRecordingRepository) FindByID(
	ctx context.Context,
	id string,
	showAll bool,
	viewerMihuaWorkNumber string,
) (*model.TelemarketingRecording, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}

	query := r.scopedQuery(
		r.db.WithContext(ctx).Model(&model.TelemarketingRecording{}),
		showAll,
		viewerMihuaWorkNumber,
	)
	if query == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var item model.TelemarketingRecording
	if err := query.Where("id = ?", id).Take(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *gormTelemarketingRecordingRepository) UpsertBatch(
	ctx context.Context,
	items []model.TelemarketingRecordingUpsertInput,
) ([]model.TelemarketingRecording, error) {
	items = dedupeTelemarketingRecordingUpsertInputs(items)
	if len(items) == 0 {
		return []model.TelemarketingRecording{}, nil
	}

	now := time.Now().UTC()
	result := make([]model.TelemarketingRecording, 0, len(items))
	for _, item := range items {
		row := model.TelemarketingRecording{
			ID:                    item.ID,
			CCNumber:              item.CCNumber,
			SID:                   item.SID,
			SeID:                  item.SeID,
			Ccgeid:                item.Ccgeid,
			CallType:              item.CallType,
			OutlineNumber:         item.OutlineNumber,
			EncryptedOutlineNum:   item.EncryptedOutlineNum,
			SwitchNumber:          item.SwitchNumber,
			Initiator:             item.Initiator,
			InitiatorCallID:       item.InitiatorCallID,
			ServiceNumber:         item.ServiceNumber,
			ServiceUID:            item.ServiceUID,
			ServiceSeatName:       item.ServiceSeatName,
			ServiceSeatWorkNumber: item.ServiceSeatWorkNumber,
			ServiceGroupName:      item.ServiceGroupName,
			InitiateTime:          item.InitiateTime,
			RingTime:              item.RingTime,
			ConfirmTime:           item.ConfirmTime,
			DisconnectTime:        item.DisconnectTime,
			ConversationTime:      item.ConversationTime,
			DurationSecond:        item.DurationSecond,
			DurationText:          item.DurationText,
			ValidDurationText:     item.ValidDurationText,
			CustomerRingDuration:  item.CustomerRingDuration,
			SeatRingDuration:      item.SeatRingDuration,
			RecordStatus:          item.RecordStatus,
			RecordFilename:        item.RecordFilename,
			RecordResToken:        item.RecordResToken,
			EvaluateValue:         item.EvaluateValue,
			CMResult:              item.CMResult,
			CMDescription:         item.CMDescription,
			Attribution:           item.Attribution,
			StopReason:            item.StopReason,
			CustomerFailReason:    item.CustomerFailReason,
			CustomerName:          item.CustomerName,
			CustomerCompany:       item.CustomerCompany,
			GroupNames:            item.GroupNames,
			SeatNames:             item.SeatNames,
			SeatNumbers:           item.SeatNumbers,
			SeatWorkNumbers:       item.SeatWorkNumbers,
			EnterpriseName:        item.EnterpriseName,
			DistrictName:          item.DistrictName,
			ServiceDeviceNumber:   item.ServiceDeviceNumber,
			CallAnswerResult:      item.CallAnswerResult,
			CallHangupParty:       item.CallHangupParty,
			MatchedUserID:         item.MatchedUserID,
			MatchedUserName:       item.MatchedUserName,
			RoleName:              item.RoleName,
			RemoteCreatedAt:       item.RemoteCreatedAt,
			RemoteUpdatedAt:       item.RemoteUpdatedAt,
			RawPayload:            item.RawPayload,
			CreatedAt:             now,
			UpdatedAt:             now,
		}

		if err := r.db.WithContext(ctx).
			Table(row.TableName()).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "cc_number"}},
				DoUpdates: clause.Assignments(map[string]any{
					"sid":                      row.SID,
					"seid":                     row.SeID,
					"ccgeid":                   row.Ccgeid,
					"call_type":                row.CallType,
					"outline_number":           row.OutlineNumber,
					"encrypted_outline_number": row.EncryptedOutlineNum,
					"switch_number":            row.SwitchNumber,
					"initiator":                row.Initiator,
					"initiator_call_id":        row.InitiatorCallID,
					"service_number":           row.ServiceNumber,
					"service_uid":              row.ServiceUID,
					"service_seat_name":        row.ServiceSeatName,
					"service_seat_worknumber":  row.ServiceSeatWorkNumber,
					"service_group_name":       row.ServiceGroupName,
					"initiate_time":            row.InitiateTime,
					"ring_time":                row.RingTime,
					"confirm_time":             row.ConfirmTime,
					"disconnect_time":          row.DisconnectTime,
					"conversation_time":        row.ConversationTime,
					"duration_second":          row.DurationSecond,
					"duration_text":            row.DurationText,
					"valid_duration_text":      row.ValidDurationText,
					"customer_ring_duration":   row.CustomerRingDuration,
					"seat_ring_duration":       row.SeatRingDuration,
					"record_status":            row.RecordStatus,
					"record_filename":          row.RecordFilename,
					"record_res_token":         row.RecordResToken,
					"evaluate_value":           row.EvaluateValue,
					"cm_result":                row.CMResult,
					"cm_description":           row.CMDescription,
					"attribution":              row.Attribution,
					"stop_reason":              row.StopReason,
					"customer_fail_reason":     row.CustomerFailReason,
					"customer_name":            row.CustomerName,
					"customer_company":         row.CustomerCompany,
					"group_names":              row.GroupNames,
					"seat_names":               row.SeatNames,
					"seat_numbers":             row.SeatNumbers,
					"seat_work_numbers":        row.SeatWorkNumbers,
					"enterprise_name":          row.EnterpriseName,
					"district_name":            row.DistrictName,
					"service_device_number":    row.ServiceDeviceNumber,
					"call_answer_result":       row.CallAnswerResult,
					"call_hangup_party":        row.CallHangupParty,
					"matched_user_id":          row.MatchedUserID,
					"matched_user_name":        row.MatchedUserName,
					"role_name":                row.RoleName,
					"remote_created_at":        row.RemoteCreatedAt,
					"remote_updated_at":        row.RemoteUpdatedAt,
					"raw_payload":              row.RawPayload,
					"updated_at":               row.UpdatedAt,
				}),
			}).
			Create(&row).Error; err != nil {
			return nil, err
		}

		var saved model.TelemarketingRecording
		if err := r.db.WithContext(ctx).
			Table(row.TableName()).
			Where("cc_number = ?", row.CCNumber).
			Take(&saved).Error; err != nil {
			return nil, err
		}
		result = append(result, saved)
	}

	return result, nil
}

func (r *gormTelemarketingRecordingRepository) ListEnabledTelemarketingUsersByWorkNumbers(
	ctx context.Context,
	workNumbers []string,
) (map[string]model.TelemarketingRecordingMatchedUser, error) {
	workNumbers = uniqueTrimmedStrings(workNumbers)
	if len(workNumbers) == 0 {
		return map[string]model.TelemarketingRecordingMatchedUser{}, nil
	}

	type telemarketingUserRow struct {
		UserID     int64  `gorm:"column:user_id"`
		Username   string `gorm:"column:username"`
		Nickname   string `gorm:"column:nickname"`
		UserName   string `gorm:"column:user_name"`
		RoleName   string `gorm:"column:role_name"`
		WorkNumber string `gorm:"column:work_number"`
	}

	var rows []telemarketingUserRow
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select(
			"u.id AS user_id",
			"u.username AS username",
			"COALESCE(NULLIF(u.nickname, ''), '') AS nickname",
			"COALESCE(NULLIF(u.nickname, ''), NULLIF(u.username, ''), '') AS user_name",
			"CASE WHEN COALESCE(r.label, '') <> '' THEN r.label ELSE COALESCE(r.name, '') END AS role_name",
			"COALESCE(NULLIF(u.mihua_work_number, ''), '') AS work_number",
		).
		Joins("JOIN roles AS r ON r.id = u.role_id").
		Where("u.status = ?", model.UserStatusEnabled).
		Where("COALESCE(NULLIF(u.mihua_work_number, ''), '') IN ?", workNumbers).
		Where("(r.name IN ? OR r.label IN ?)", salesDailyScoreTelemarketingRoleNames, salesDailyScoreTelemarketingRoleLabels).
		Order("u.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]model.TelemarketingRecordingMatchedUser, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.WorkNumber)
		if key == "" {
			continue
		}
		result[key] = model.TelemarketingRecordingMatchedUser{
			UserID:     row.UserID,
			Username:   strings.TrimSpace(row.Username),
			Nickname:   strings.TrimSpace(row.Nickname),
			UserName:   strings.TrimSpace(row.UserName),
			RoleName:   strings.TrimSpace(row.RoleName),
			WorkNumber: key,
		}
	}
	return result, nil
}

func (r *gormTelemarketingRecordingRepository) scopedQuery(
	query *gorm.DB,
	showAll bool,
	viewerMihuaWorkNumber string,
) *gorm.DB {
	if showAll {
		return query
	}
	viewerMihuaWorkNumber = strings.TrimSpace(viewerMihuaWorkNumber)
	if viewerMihuaWorkNumber == "" {
		return nil
	}
	return query.Where("service_seat_worknumber = ?", viewerMihuaWorkNumber)
}

func dedupeTelemarketingRecordingUpsertInputs(items []model.TelemarketingRecordingUpsertInput) []model.TelemarketingRecordingUpsertInput {
	if len(items) == 0 {
		return []model.TelemarketingRecordingUpsertInput{}
	}

	order := make([]string, 0, len(items))
	merged := make(map[string]model.TelemarketingRecordingUpsertInput, len(items))
	usedIDs := make(map[string]string, len(items))
	for _, item := range items {
		item = normalizeTelemarketingRecordingUpsertInput(item, usedIDs)
		if item.ID == "" || item.CCNumber == "" {
			continue
		}
		if _, exists := merged[item.CCNumber]; !exists {
			order = append(order, item.CCNumber)
		}
		merged[item.CCNumber] = item
	}

	result := make([]model.TelemarketingRecordingUpsertInput, 0, len(order))
	for _, key := range order {
		result = append(result, merged[key])
	}
	return result
}

func normalizeTelemarketingRecordingUpsertInput(
	item model.TelemarketingRecordingUpsertInput,
	usedIDs map[string]string,
) model.TelemarketingRecordingUpsertInput {
	item.ID = strings.TrimSpace(item.ID)
	item.CCNumber = strings.TrimSpace(item.CCNumber)
	if item.CCNumber == "" {
		item.ID = ""
		return item
	}
	item.ID = ensureUniqueTelemarketingRecordingID(item.ID, item.CCNumber, usedIDs)
	return item
}

func ensureUniqueTelemarketingRecordingID(preferredID, ccNumber string, usedIDs map[string]string) string {
	candidates := []string{
		sanitizeTelemarketingRecordingID(preferredID),
		buildTelemarketingRecordingFallbackID(ccNumber),
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		existingCCNumber, exists := usedIDs[candidate]
		if exists && existingCCNumber != ccNumber {
			continue
		}
		usedIDs[candidate] = ccNumber
		return candidate
	}

	sum := sha1.Sum([]byte(strings.TrimSpace(ccNumber) + "|" + strings.TrimSpace(preferredID)))
	candidate := "cc_" + hex.EncodeToString(sum[:])
	usedIDs[candidate] = ccNumber
	return candidate
}

func sanitizeTelemarketingRecordingID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= telemarketingRecordingIDMaxLen {
		return value
	}

	sum := sha1.Sum([]byte(value))
	return "id_" + hex.EncodeToString(sum[:])
}

func buildTelemarketingRecordingFallbackID(ccNumber string) string {
	ccNumber = strings.TrimSpace(ccNumber)
	if ccNumber == "" {
		return ""
	}
	if len(ccNumber) <= telemarketingRecordingIDMaxLen {
		return ccNumber
	}

	sum := sha1.Sum([]byte(ccNumber))
	return "cc_" + hex.EncodeToString(sum[:])
}

func telemarketingRecordingUnixRange(startDate, endDate string) (int64, int64) {
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" && endDate == "" {
		return 0, 0
	}
	if startDate == "" {
		startDate = endDate
	}
	if endDate == "" {
		endDate = startDate
	}

	startValue, err := time.ParseInLocation("2006-01-02", startDate, time.Local)
	if err != nil {
		return 0, 0
	}
	endValue, err := time.ParseInLocation("2006-01-02", endDate, time.Local)
	if err != nil {
		return 0, 0
	}
	if endValue.Before(startValue) {
		startValue, endValue = endValue, startValue
	}
	return startValue.Unix(), endValue.AddDate(0, 0, 1).Unix()
}
