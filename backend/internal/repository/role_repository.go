package repository

import (
	"backend/internal/model"
	"context"

	"gorm.io/gorm"
)

type RoleRepository interface {
	List(ctx context.Context) ([]model.Role, error)
	FindByID(ctx context.Context, id int64) (*model.Role, error)
	FindByName(ctx context.Context, name string) (*model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id int64) error
}

type gormRoleRepository struct{ db *gorm.DB }

func NewGormRoleRepository(db *gorm.DB) RoleRepository {
	return &gormRoleRepository{db: db}
}

func NewSQLiteRoleRepository(db *gorm.DB) RoleRepository {
	return NewGormRoleRepository(db)
}

func (r *gormRoleRepository) List(ctx context.Context) ([]model.Role, error) {
	var list []model.Role
	err := r.db.WithContext(ctx).
		Table("roles").
		Select("id", "name", "label", "sort", "created_at").
		Order("sort, id").
		Scan(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Role{}
	}
	return list, nil
}

func (r *gormRoleRepository) FindByID(ctx context.Context, id int64) (*model.Role, error) {
	role := &model.Role{}
	err := r.db.WithContext(ctx).
		Table("roles").
		Select("id", "name", "label", "sort", "created_at").
		Where("id = ?", id).
		Take(role).Error
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *gormRoleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	role := &model.Role{}
	err := r.db.WithContext(ctx).
		Table("roles").
		Select("id", "name", "label", "sort", "created_at").
		Where("name = ?", name).
		Take(role).Error
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *gormRoleRepository) Create(ctx context.Context, role *model.Role) error {
	type roleRow struct {
		ID        int64  `gorm:"column:id;primaryKey;autoIncrement"`
		Name      string `gorm:"column:name"`
		Label     string `gorm:"column:label"`
		Sort      int    `gorm:"column:sort"`
		CreatedAt any    `gorm:"column:created_at"`
	}

	row := roleRow{Name: role.Name, Label: role.Label, Sort: role.Sort}
	if err := r.db.WithContext(ctx).Table("roles").Create(&row).Error; err != nil {
		return err
	}
	role.ID = row.ID
	return r.db.WithContext(ctx).
		Table("roles").
		Select("id", "name", "label", "sort", "created_at").
		Where("id = ?", role.ID).
		Take(role).Error
}

func (r *gormRoleRepository) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).
		Table("roles").
		Where("id = ?", role.ID).
		Updates(map[string]interface{}{
			"name":  role.Name,
			"label": role.Label,
			"sort":  role.Sort,
		}).Error
}

func (r *gormRoleRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Table("roles").
		Where("id = ?", id).
		Delete(nil).Error
}
