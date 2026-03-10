package repository

import (
	"backend/internal/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id int64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	ListWithRole(ctx context.Context) ([]model.UserWithRole, error)
	SearchWithRole(ctx context.Context, keyword string) ([]model.UserWithRole, error)
	Update(ctx context.Context, user *model.User) error
	BatchUpdateStatus(ctx context.Context, ids []int64, status string) (int64, error)
	Delete(ctx context.Context, id int64) error
}

type gormUserRepository struct{ db *gorm.DB }

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func NewSQLiteUserRepository(db *gorm.DB) UserRepository {
	return NewGormUserRepository(db)
}

func (r *gormUserRepository) Create(ctx context.Context, user *model.User) error {
	type userRow struct {
		ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
		Username  string    `gorm:"column:username"`
		Password  string    `gorm:"column:password"`
		Salt      string    `gorm:"column:salt"`
		Nickname  string    `gorm:"column:nickname"`
		Email     string    `gorm:"column:email"`
		Mobile    string    `gorm:"column:mobile"`
		Avatar    string    `gorm:"column:avatar"`
		RoleID    int64     `gorm:"column:role_id"`
		ParentID  *int64    `gorm:"column:parent_id"`
		Status    string    `gorm:"column:status"`
		CreatedAt time.Time `gorm:"column:created_at"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}

	now := time.Now().UTC()
	row := userRow{
		Username:  user.Username,
		Password:  user.Password,
		Salt:      user.Salt,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Mobile:    user.Mobile,
		Avatar:    user.Avatar,
		RoleID:    user.RoleID,
		ParentID:  user.ParentID,
		Status:    user.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.db.WithContext(ctx).Table("users").Create(&row).Error; err != nil {
		return err
	}
	user.ID = row.ID
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *gormUserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	u := &model.User{}
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id", "username", "password", "salt", "nickname", "email", "mobile", "avatar", "role_id", "parent_id", "status", "created_at", "updated_at").
		Where("id = ?", id).
		Take(u).Error
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *gormUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	u := &model.User{}
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id", "username", "password", "salt", "nickname", "email", "mobile", "avatar", "role_id", "parent_id", "status", "created_at", "updated_at").
		Where("username = ?", username).
		Take(u).Error
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *gormUserRepository) ListWithRole(ctx context.Context) ([]model.UserWithRole, error) {
	var list []model.UserWithRole
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select(
			"u.id", "u.username", "u.nickname", "u.email", "u.mobile", "u.avatar",
			"u.role_id", "u.parent_id", "u.status", "u.created_at", "u.updated_at",
			"COALESCE(r.name, '') AS role_name", "COALESCE(r.label, '') AS role_label",
		).
		Joins("LEFT JOIN roles r ON u.role_id = r.id").
		Order("u.id").
		Scan(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.UserWithRole{}
	}
	return list, nil
}

func (r *gormUserRepository) SearchWithRole(ctx context.Context, keyword string) ([]model.UserWithRole, error) {
	searchPattern := "%" + keyword + "%"
	var list []model.UserWithRole
	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select(
			"u.id", "u.username", "u.nickname", "u.email", "u.mobile", "u.avatar",
			"u.role_id", "u.parent_id", "u.status", "u.created_at", "u.updated_at",
			"COALESCE(r.name, '') AS role_name", "COALESCE(r.label, '') AS role_label",
		).
		Joins("LEFT JOIN roles r ON u.role_id = r.id").
		Where(
			"u.username LIKE ? OR u.nickname LIKE ? OR u.email LIKE ? OR u.mobile LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		).
		Order("u.id").
		Scan(&list).Error
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.UserWithRole{}
	}
	return list, nil
}

func (r *gormUserRepository) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).
		Table("users").
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"username":   user.Username,
			"password":   user.Password,
			"nickname":   user.Nickname,
			"email":      user.Email,
			"mobile":     user.Mobile,
			"avatar":     user.Avatar,
			"role_id":    user.RoleID,
			"parent_id":  user.ParentID,
			"status":     user.Status,
			"updated_at": user.UpdatedAt,
		}).Error
}

func (r *gormUserRepository) BatchUpdateStatus(ctx context.Context, ids []int64, status string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := r.db.WithContext(ctx).
		Table("users").
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *gormUserRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Table("users").
		Where("id = ?", id).
		Delete(nil).Error
}
