package repository

import (
	"backend/internal/model"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrHanghangCRMUserNotMatched = errors.New("hanghang crm user not matched")

type HanghangCRMDailyUserCallStatRepository interface {
	UpsertBatch(ctx context.Context, items []model.DailyUserCallStatUpsertInput) ([]model.DailyUserCallStat, error)
	FindUserIDByNicknameAndHanghangCRMMobile(ctx context.Context, nickname, hanghangCRMMobile string) (*int64, error)
}

type gormHanghangCRMDailyUserCallStatRepository struct {
	db *gorm.DB
}

type dailyUserCallStatRow struct {
	ID                  int64     `gorm:"column:id;primaryKey;autoIncrement"`
	StatDate            string    `gorm:"column:stat_date"`
	UserID              *int64    `gorm:"column:user_id"`
	RealName            string    `gorm:"column:real_name"`
	Mobile              string    `gorm:"column:mobile"`
	BindNum             int       `gorm:"column:bind_num"`
	CallNum             int       `gorm:"column:call_num"`
	NotConnected        int       `gorm:"column:not_connected"`
	ConnectionRate      float64   `gorm:"column:connection_rate"`
	TimeTotal           int       `gorm:"column:time_total"`
	TotalMinute         string    `gorm:"column:total_minute"`
	TotalSecond         int       `gorm:"column:total_second"`
	AverageCallDuration float64   `gorm:"column:average_call_duration"`
	AverageCallSecond   float64   `gorm:"column:average_call_second"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func NewGormHanghangCRMDailyUserCallStatRepository(db *gorm.DB) HanghangCRMDailyUserCallStatRepository {
	return &gormHanghangCRMDailyUserCallStatRepository{db: db}
}

func (r *gormHanghangCRMDailyUserCallStatRepository) UpsertBatch(
	ctx context.Context,
	items []model.DailyUserCallStatUpsertInput,
) ([]model.DailyUserCallStat, error) {
	items = dedupeDailyUserCallStatUpsertInputs(items)
	if len(items) == 0 {
		return []model.DailyUserCallStat{}, nil
	}

	now := time.Now().UTC()
	result := make([]model.DailyUserCallStat, 0, len(items))
	for _, item := range items {
		row := dailyUserCallStatRow{
			StatDate:            item.StatDate,
			UserID:              item.UserID,
			RealName:            item.RealName,
			Mobile:              item.Mobile,
			BindNum:             item.BindNum,
			CallNum:             item.CallNum,
			NotConnected:        item.NotConnected,
			ConnectionRate:      item.ConnectionRate,
			TimeTotal:           item.TimeTotal,
			TotalMinute:         item.TotalMinute,
			TotalSecond:         item.TotalSecond,
			AverageCallDuration: item.AverageCallDuration,
			AverageCallSecond:   item.AverageCallSecond,
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		err := r.db.WithContext(ctx).Table("daily_user_call_stats").Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "stat_date"},
				{Name: "real_name"},
				{Name: "mobile"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"user_id":               row.UserID,
				"bind_num":              row.BindNum,
				"call_num":              row.CallNum,
				"not_connected":         row.NotConnected,
				"connection_rate":       row.ConnectionRate,
				"time_total":            row.TimeTotal,
				"total_minute":          row.TotalMinute,
				"total_second":          row.TotalSecond,
				"average_call_duration": row.AverageCallDuration,
				"average_call_second":   row.AverageCallSecond,
				"updated_at":            row.UpdatedAt,
			}),
		}).Create(&row).Error
		if err != nil {
			return nil, err
		}

		var saved dailyUserCallStatRow
		if err := r.db.WithContext(ctx).
			Table("daily_user_call_stats").
			Where("stat_date = ? AND real_name = ? AND mobile = ?", row.StatDate, row.RealName, row.Mobile).
			Take(&saved).Error; err != nil {
			return nil, err
		}
		result = append(result, mapDailyUserCallStatRowToModel(saved))
	}

	return result, nil
}

func dedupeDailyUserCallStatUpsertInputs(
	items []model.DailyUserCallStatUpsertInput,
) []model.DailyUserCallStatUpsertInput {
	if len(items) == 0 {
		return []model.DailyUserCallStatUpsertInput{}
	}

	order := make([]string, 0, len(items))
	merged := make(map[string]model.DailyUserCallStatUpsertInput, len(items))
	for _, item := range items {
		key := fmt.Sprintf(
			"%s\x00%s\x00%s",
			strings.TrimSpace(item.StatDate),
			strings.TrimSpace(item.RealName),
			strings.TrimSpace(item.Mobile),
		)
		if _, exists := merged[key]; !exists {
			order = append(order, key)
		}
		merged[key] = model.DailyUserCallStatUpsertInput{
			StatDate:            strings.TrimSpace(item.StatDate),
			UserID:              item.UserID,
			RealName:            strings.TrimSpace(item.RealName),
			Mobile:              strings.TrimSpace(item.Mobile),
			BindNum:             item.BindNum,
			CallNum:             item.CallNum,
			NotConnected:        item.NotConnected,
			ConnectionRate:      item.ConnectionRate,
			TimeTotal:           item.TimeTotal,
			TotalMinute:         strings.TrimSpace(item.TotalMinute),
			TotalSecond:         item.TotalSecond,
			AverageCallDuration: item.AverageCallDuration,
			AverageCallSecond:   item.AverageCallSecond,
		}
	}

	result := make([]model.DailyUserCallStatUpsertInput, 0, len(order))
	for _, key := range order {
		result = append(result, merged[key])
	}
	return result
}

func (r *gormHanghangCRMDailyUserCallStatRepository) FindUserIDByNicknameAndHanghangCRMMobile(
	ctx context.Context,
	nickname, hanghangCRMMobile string,
) (*int64, error) {
	nickname = strings.TrimSpace(nickname)
	hanghangCRMMobile = strings.TrimSpace(hanghangCRMMobile)

	if hanghangCRMMobile != "" {
		userID, err := r.findEnabledUserIDByHanghangCRMMobile(ctx, hanghangCRMMobile)
		if err != nil {
			return nil, err
		}
		if userID != nil {
			return userID, nil
		}
	}

	if nickname == "" {
		return nil, nil
	}

	return r.findEnabledUserIDByNickname(ctx, nickname, "")
}

func (r *gormHanghangCRMDailyUserCallStatRepository) findEnabledUserIDByHanghangCRMMobile(
	ctx context.Context,
	hanghangCRMMobile string,
) (*int64, error) {
	return r.findEnabledUserIDByLegacyHanghangCRMMobile(ctx, hanghangCRMMobile)
}

func (r *gormHanghangCRMDailyUserCallStatRepository) findEnabledUserIDByLegacyHanghangCRMMobile(
	ctx context.Context,
	hanghangCRMMobile string,
) (*int64, error) {
	type userIDRow struct {
		ID int64 `gorm:"column:id"`
	}

	var row userIDRow
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id").
		Where("status = ?", model.UserStatusEnabled).
		Where("hanghang_crm_mobile = ?", hanghangCRMMobile).
		Order("id ASC").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row.ID, nil
}

func (r *gormHanghangCRMDailyUserCallStatRepository) findEnabledUserIDByNickname(
	ctx context.Context,
	nickname, hanghangCRMMobile string,
) (*int64, error) {
	type userIDRow struct {
		ID int64 `gorm:"column:id"`
	}

	query := r.db.WithContext(ctx).
		Table("users").
		Select("id").
		Where("status = ?", model.UserStatusEnabled).
		Where("nickname = ?", nickname)
	if hanghangCRMMobile != "" {
		query = query.Where("hanghang_crm_mobile = ?", hanghangCRMMobile)
	}

	var row userIDRow
	err := query.
		Order("id ASC").
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &row.ID, nil
}

func mapDailyUserCallStatRowToModel(row dailyUserCallStatRow) model.DailyUserCallStat {
	return model.DailyUserCallStat{
		ID:                  row.ID,
		StatDate:            row.StatDate,
		UserID:              row.UserID,
		RealName:            row.RealName,
		Mobile:              row.Mobile,
		BindNum:             row.BindNum,
		CallNum:             row.CallNum,
		NotConnected:        row.NotConnected,
		ConnectionRate:      row.ConnectionRate,
		TimeTotal:           row.TimeTotal,
		TotalMinute:         row.TotalMinute,
		TotalSecond:         row.TotalSecond,
		AverageCallDuration: row.AverageCallDuration,
		AverageCallSecond:   row.AverageCallSecond,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}
