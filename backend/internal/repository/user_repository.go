package repository

import (
	"backend/internal/model"
	"context"
	"strings"
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

func dedupeMobiles(mobiles []string) []string {
	result := make([]string, 0, len(mobiles))
	seen := make(map[string]struct{}, len(mobiles))
	for _, mobile := range mobiles {
		trimmed := strings.TrimSpace(mobile)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func firstHanghangCRMMobile(mobiles []string, fallback string) string {
	cleaned := dedupeMobiles(mobiles)
	if len(cleaned) > 0 {
		return cleaned[0]
	}
	return strings.TrimSpace(fallback)
}

func listUserHanghangCRMMobiles(ctx context.Context, db *gorm.DB, userIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	type mobileRow struct {
		UserID int64  `gorm:"column:user_id"`
		Mobile string `gorm:"column:mobile"`
	}

	var rows []mobileRow
	if err := db.WithContext(ctx).
		Table("user_hanghang_crm_mobiles").
		Select("user_id", "mobile").
		Where("user_id IN ?", userIDs).
		Order("user_id ASC, is_primary DESC, id ASC").
		Scan(&rows).Error; err != nil {
		lowered := strings.ToLower(err.Error())
		if strings.Contains(lowered, "no such table") || strings.Contains(lowered, "doesn't exist") {
			return result, nil
		}
		return nil, err
	}

	for _, row := range rows {
		result[row.UserID] = append(result[row.UserID], strings.TrimSpace(row.Mobile))
	}
	return result, nil
}

func syncUserHanghangCRMMobiles(ctx context.Context, tx *gorm.DB, userID int64, mobiles []string) error {
	if err := tx.WithContext(ctx).
		Table("user_hanghang_crm_mobiles").
		Where("user_id = ?", userID).
		Delete(nil).Error; err != nil {
		return err
	}

	cleaned := dedupeMobiles(mobiles)
	if len(cleaned) == 0 {
		return nil
	}

	rows := make([]map[string]interface{}, 0, len(cleaned))
	now := time.Now().UTC()
	for idx, mobile := range cleaned {
		rows = append(rows, map[string]interface{}{
			"user_id":    userID,
			"mobile":     mobile,
			"is_primary": idx == 0,
			"created_at": now,
			"updated_at": now,
		})
	}
	return tx.WithContext(ctx).Table("user_hanghang_crm_mobiles").Create(&rows).Error
}

func (r *gormUserRepository) populateUserHanghangCRMMobiles(ctx context.Context, user *model.User) error {
	if user == nil || user.ID <= 0 {
		return nil
	}
	mobiles, err := listUserHanghangCRMMobiles(ctx, r.db, []int64{user.ID})
	if err != nil {
		return err
	}
	user.HanghangCRMMobiles = mobiles[user.ID]
	user.HanghangCRMMobile = firstHanghangCRMMobile(user.HanghangCRMMobiles, user.HanghangCRMMobile)
	return nil
}

func (r *gormUserRepository) populateUserWithRoleHanghangCRMMobiles(ctx context.Context, users []model.UserWithRole) error {
	if len(users) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(users))
	for _, user := range users {
		if user.ID > 0 {
			ids = append(ids, user.ID)
		}
	}
	mobiles, err := listUserHanghangCRMMobiles(ctx, r.db, ids)
	if err != nil {
		return err
	}
	for idx := range users {
		users[idx].HanghangCRMMobiles = mobiles[users[idx].ID]
		users[idx].HanghangCRMMobile = firstHanghangCRMMobile(users[idx].HanghangCRMMobiles, users[idx].HanghangCRMMobile)
	}
	return nil
}

func (r *gormUserRepository) Create(ctx context.Context, user *model.User) error {
	type userRow struct {
		ID                int64     `gorm:"column:id;primaryKey;autoIncrement"`
		Username          string    `gorm:"column:username"`
		Password          string    `gorm:"column:password"`
		Salt              string    `gorm:"column:salt"`
		Nickname          string    `gorm:"column:nickname"`
		Email             string    `gorm:"column:email"`
		Mobile            string    `gorm:"column:mobile"`
		HanghangCRMMobile string    `gorm:"column:hanghang_crm_mobile"`
		Avatar            string    `gorm:"column:avatar"`
		RoleID            int64     `gorm:"column:role_id"`
		ParentID          *int64    `gorm:"column:parent_id"`
		Status            string    `gorm:"column:status"`
		CreatedAt         time.Time `gorm:"column:created_at"`
		UpdatedAt         time.Time `gorm:"column:updated_at"`
	}

	now := time.Now().UTC()
	row := userRow{
		Username:          user.Username,
		Password:          user.Password,
		Salt:              user.Salt,
		Nickname:          user.Nickname,
		Email:             user.Email,
		Mobile:            user.Mobile,
		HanghangCRMMobile: strings.TrimSpace(user.HanghangCRMMobile),
		Avatar:            user.Avatar,
		RoleID:            user.RoleID,
		ParentID:          user.ParentID,
		Status:            user.Status,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := r.db.WithContext(ctx).Table("users").Create(&row).Error; err != nil {
		return err
	}
	user.ID = row.ID
	user.HanghangCRMMobile = row.HanghangCRMMobile
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *gormUserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	u := &model.User{}
	err := r.db.WithContext(ctx).
		Table("users").
		Select("id", "username", "password", "salt", "nickname", "email", "mobile", "hanghang_crm_mobile", "avatar", "role_id", "parent_id", "status", "created_at", "updated_at").
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
		Select("id", "username", "password", "salt", "nickname", "email", "mobile", "hanghang_crm_mobile", "avatar", "role_id", "parent_id", "status", "created_at", "updated_at").
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
			"u.id", "u.username", "u.nickname", "u.email", "u.mobile", "u.hanghang_crm_mobile", "u.avatar",
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
			"u.id", "u.username", "u.nickname", "u.email", "u.mobile", "u.hanghang_crm_mobile", "u.avatar",
			"u.role_id", "u.parent_id", "u.status", "u.created_at", "u.updated_at",
			"COALESCE(r.name, '') AS role_name", "COALESCE(r.label, '') AS role_label",
		).
		Joins("LEFT JOIN roles r ON u.role_id = r.id").
		Where(
			`u.username LIKE ? OR u.nickname LIKE ? OR u.email LIKE ? OR u.mobile LIKE ? OR u.hanghang_crm_mobile LIKE ?`,
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
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
	user.HanghangCRMMobile = strings.TrimSpace(user.HanghangCRMMobile)
	return r.db.WithContext(ctx).
		Table("users").
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"username":            user.Username,
			"password":            user.Password,
			"nickname":            user.Nickname,
			"email":               user.Email,
			"mobile":              user.Mobile,
			"hanghang_crm_mobile": user.HanghangCRMMobile,
			"avatar":              user.Avatar,
			"role_id":             user.RoleID,
			"parent_id":           user.ParentID,
			"status":              user.Status,
			"updated_at":          user.UpdatedAt,
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).
			Table("user_hanghang_crm_mobiles").
			Where("user_id = ?", id).
			Delete(nil).Error; err != nil {
			return err
		}
		return tx.WithContext(ctx).
			Table("users").
			Where("id = ?", id).
			Delete(nil).Error
	})
}
