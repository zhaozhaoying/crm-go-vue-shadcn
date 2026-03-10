package repository

import (
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrResourcePoolNotFound         = errors.New("resource pool item not found")
	ErrResourcePoolAlreadyConverted = errors.New("resource pool item already converted")
)

type ResourcePoolRepository interface {
	List(ctx context.Context, filter model.ResourcePoolListFilter) (model.ResourcePoolListResult, error)
	UpsertBatch(ctx context.Context, items []model.ResourcePoolItemUpsertInput) ([]model.ResourcePoolItem, error)
	GetByID(ctx context.Context, id int64) (*model.ResourcePoolItem, error)
	MarkConverted(ctx context.Context, id int64, customerID int64, operatorUserID int64) error
}

type gormResourcePoolRepository struct {
	db *gorm.DB
}

type resourcePoolRow struct {
	ID                  int64         `gorm:"column:id;primaryKey;autoIncrement"`
	Name                string        `gorm:"column:name"`
	Phone               string        `gorm:"column:phone"`
	Address             string        `gorm:"column:address"`
	Province            string        `gorm:"column:province"`
	City                string        `gorm:"column:city"`
	Area                string        `gorm:"column:area"`
	Latitude            float64       `gorm:"column:latitude"`
	Longitude           float64       `gorm:"column:longitude"`
	Source              string        `gorm:"column:source"`
	SourceUID           string        `gorm:"column:source_uid"`
	SearchKeyword       string        `gorm:"column:search_keyword"`
	SearchRadius        int           `gorm:"column:search_radius"`
	SearchRegion        string        `gorm:"column:search_region"`
	QueryAddress        string        `gorm:"column:query_address"`
	CenterLatitude      float64       `gorm:"column:center_latitude"`
	CenterLongitude     float64       `gorm:"column:center_longitude"`
	CreatedBy           int64         `gorm:"column:created_by"`
	Converted           int           `gorm:"column:converted"`
	ConvertedCustomerID sql.NullInt64 `gorm:"column:converted_customer_id"`
	ConvertedAt         *time.Time    `gorm:"column:converted_at"`
	ConvertedBy         sql.NullInt64 `gorm:"column:converted_by"`
	CreatedAt           time.Time     `gorm:"column:created_at"`
	UpdatedAt           time.Time     `gorm:"column:updated_at"`
}

func NewGormResourcePoolRepository(db *gorm.DB) ResourcePoolRepository {
	return &gormResourcePoolRepository{db: db}
}

func (r *gormResourcePoolRepository) List(ctx context.Context, filter model.ResourcePoolListFilter) (model.ResourcePoolListResult, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}

	base := r.db.WithContext(ctx).Table("resource_pool")
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		base = base.Where("name LIKE ? OR phone LIKE ? OR address LIKE ?", pattern, pattern, pattern)
	}
	if filter.HasPhone != nil {
		if *filter.HasPhone {
			base = base.Where("TRIM(COALESCE(phone, '')) <> ''")
		} else {
			base = base.Where("TRIM(COALESCE(phone, '')) = ''")
		}
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return model.ResourcePoolListResult{}, err
	}

	var rows []resourcePoolRow
	err := base.Session(&gorm.Session{}).
		Order("updated_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error
	if err != nil {
		return model.ResourcePoolListResult{}, err
	}

	items := make([]model.ResourcePoolItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapResourcePoolRowToModel(row))
	}
	if items == nil {
		items = []model.ResourcePoolItem{}
	}

	return model.ResourcePoolListResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (r *gormResourcePoolRepository) UpsertBatch(ctx context.Context, items []model.ResourcePoolItemUpsertInput) ([]model.ResourcePoolItem, error) {
	if len(items) == 0 {
		return []model.ResourcePoolItem{}, nil
	}

	now := time.Now().UTC()
	result := make([]model.ResourcePoolItem, 0, len(items))
	for _, item := range items {
		row := resourcePoolRow{
			Name:            item.Name,
			Phone:           item.Phone,
			Address:         item.Address,
			Province:        item.Province,
			City:            item.City,
			Area:            item.Area,
			Latitude:        item.Latitude,
			Longitude:       item.Longitude,
			Source:          item.Source,
			SourceUID:       item.SourceUID,
			SearchKeyword:   item.SearchKeyword,
			SearchRadius:    item.SearchRadius,
			SearchRegion:    item.SearchRegion,
			QueryAddress:    item.QueryAddress,
			CenterLatitude:  item.CenterLatitude,
			CenterLongitude: item.CenterLongitude,
			CreatedBy:       item.CreatedBy,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		err := r.db.WithContext(ctx).Table("resource_pool").Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "source"}, {Name: "source_uid"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"name":             row.Name,
				"phone":            row.Phone,
				"address":          row.Address,
				"province":         row.Province,
				"city":             row.City,
				"area":             row.Area,
				"latitude":         row.Latitude,
				"longitude":        row.Longitude,
				"search_keyword":   row.SearchKeyword,
				"search_radius":    row.SearchRadius,
				"search_region":    row.SearchRegion,
				"query_address":    row.QueryAddress,
				"center_latitude":  row.CenterLatitude,
				"center_longitude": row.CenterLongitude,
				"updated_at":       row.UpdatedAt,
			}),
		}).Create(&row).Error
		if err != nil {
			return nil, err
		}

		var saved resourcePoolRow
		if err := r.db.WithContext(ctx).Table("resource_pool").
			Where("source = ? AND source_uid = ?", row.Source, row.SourceUID).
			Take(&saved).Error; err != nil {
			return nil, err
		}
		result = append(result, mapResourcePoolRowToModel(saved))
	}
	return result, nil
}

func (r *gormResourcePoolRepository) GetByID(ctx context.Context, id int64) (*model.ResourcePoolItem, error) {
	var row resourcePoolRow
	err := r.db.WithContext(ctx).Table("resource_pool").Where("id = ?", id).Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResourcePoolNotFound
		}
		return nil, err
	}
	item := mapResourcePoolRowToModel(row)
	return &item, nil
}

func (r *gormResourcePoolRepository) MarkConverted(ctx context.Context, id int64, customerID int64, operatorUserID int64) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).
		Table("resource_pool").
		Where("id = ? AND COALESCE(converted, 0) = 0", id).
		Updates(map[string]interface{}{
			"converted":             1,
			"converted_customer_id": customerID,
			"converted_at":          now,
			"converted_by":          operatorUserID,
			"updated_at":            now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	type stateRow struct {
		ID        int64 `gorm:"column:id"`
		Converted int   `gorm:"column:converted"`
	}
	var row stateRow
	if err := r.db.WithContext(ctx).
		Table("resource_pool").
		Select("id", "COALESCE(converted, 0) AS converted").
		Where("id = ?", id).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResourcePoolNotFound
		}
		return err
	}
	if row.Converted == 1 {
		return ErrResourcePoolAlreadyConverted
	}
	return nil
}

func mapResourcePoolRowToModel(row resourcePoolRow) model.ResourcePoolItem {
	var convertedCustomerID *int64
	if row.ConvertedCustomerID.Valid {
		id := row.ConvertedCustomerID.Int64
		convertedCustomerID = &id
	}

	var convertedBy *int64
	if row.ConvertedBy.Valid {
		userID := row.ConvertedBy.Int64
		convertedBy = &userID
	}

	return model.ResourcePoolItem{
		ID:                  row.ID,
		Name:                row.Name,
		Phone:               row.Phone,
		Address:             row.Address,
		Province:            row.Province,
		City:                row.City,
		Area:                row.Area,
		Latitude:            row.Latitude,
		Longitude:           row.Longitude,
		Source:              row.Source,
		SourceUID:           row.SourceUID,
		SearchKeyword:       row.SearchKeyword,
		SearchRadius:        row.SearchRadius,
		SearchRegion:        row.SearchRegion,
		QueryAddress:        row.QueryAddress,
		CenterLatitude:      row.CenterLatitude,
		CenterLongitude:     row.CenterLongitude,
		CreatedBy:           row.CreatedBy,
		Converted:           row.Converted == 1,
		ConvertedCustomerID: convertedCustomerID,
		ConvertedAt:         row.ConvertedAt,
		ConvertedBy:         convertedBy,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}
