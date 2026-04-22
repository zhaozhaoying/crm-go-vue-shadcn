package repository

import (
	"backend/internal/model"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SystemSettingRepository struct {
	db *gorm.DB
}

func NewSystemSettingRepository(db *gorm.DB) *SystemSettingRepository {
	return &SystemSettingRepository{db: db}
}

func (r *SystemSettingRepository) GetSetting(key string) (*model.SystemSetting, error) {
	var s model.SystemSetting
	err := r.db.Table("system_settings").
		Select("id", "key", "value", "description", "updated_at").
		Where("`key` = ?", key).
		Take(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SystemSettingRepository) UpsertSetting(key, value, description string) error {
	type systemSettingRow struct {
		ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
		Key         string    `gorm:"column:key"`
		Value       string    `gorm:"column:value"`
		Description string    `gorm:"column:description"`
		UpdatedAt   time.Time `gorm:"column:updated_at"`
	}

	now := time.Now().UTC()
	row := systemSettingRow{
		Key:         key,
		Value:       value,
		Description: description,
		UpdatedAt:   now,
	}
	return r.db.Table("system_settings").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"value": value, "updated_at": now}),
	}).Create(&row).Error
}

func (r *SystemSettingRepository) GetAllCustomerLevels() ([]model.CustomerLevel, error) {
	var levels []model.CustomerLevel
	err := r.db.Table("customer_levels").
		Select("id", "name", "sort", "created_at").
		Order("sort ASC, id ASC").
		Find(&levels).Error
	if err != nil {
		return nil, err
	}
	if levels == nil {
		levels = []model.CustomerLevel{}
	}
	return levels, nil
}

func (r *SystemSettingRepository) CreateCustomerLevel(name string, sort int) (*model.CustomerLevel, error) {
	type customerLevelRow struct {
		ID        int       `gorm:"column:id;primaryKey;autoIncrement"`
		Name      string    `gorm:"column:name"`
		Sort      int       `gorm:"column:sort"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	row := customerLevelRow{Name: name, Sort: sort}
	if err := r.db.Table("customer_levels").Create(&row).Error; err != nil {
		return nil, err
	}
	return &model.CustomerLevel{ID: row.ID, Name: row.Name, Sort: row.Sort, CreatedAt: row.CreatedAt}, nil
}

func (r *SystemSettingRepository) UpdateCustomerLevel(id int, name string, sort int) error {
	return r.db.Table("customer_levels").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name": name,
			"sort": sort,
		}).Error
}

func (r *SystemSettingRepository) DeleteCustomerLevel(id int) error {
	return r.db.Table("customer_levels").Where("id = ?", id).Delete(nil).Error
}

func (r *SystemSettingRepository) GetAllCustomerSources() ([]model.CustomerSource, error) {
	var sources []model.CustomerSource
	err := r.db.Table("customer_sources").
		Select("id", "name", "sort", "created_at").
		Order("sort ASC, id ASC").
		Find(&sources).Error
	if err != nil {
		return nil, err
	}
	if sources == nil {
		sources = []model.CustomerSource{}
	}
	return sources, nil
}

func (r *SystemSettingRepository) CreateCustomerSource(name string, sort int) (*model.CustomerSource, error) {
	type customerSourceRow struct {
		ID        int       `gorm:"column:id;primaryKey;autoIncrement"`
		Name      string    `gorm:"column:name"`
		Sort      int       `gorm:"column:sort"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	row := customerSourceRow{Name: name, Sort: sort}
	if err := r.db.Table("customer_sources").Create(&row).Error; err != nil {
		return nil, err
	}
	return &model.CustomerSource{ID: row.ID, Name: row.Name, Sort: row.Sort, CreatedAt: row.CreatedAt}, nil
}

func (r *SystemSettingRepository) UpdateCustomerSource(id int, name string, sort int) error {
	return r.db.Table("customer_sources").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name": name,
			"sort": sort,
		}).Error
}

func (r *SystemSettingRepository) DeleteCustomerSource(id int) error {
	return r.db.Table("customer_sources").Where("id = ?", id).Delete(nil).Error
}
