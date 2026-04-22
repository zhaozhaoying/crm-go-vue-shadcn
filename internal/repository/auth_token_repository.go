package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RefreshTokenRecord struct {
	TokenHash      string `gorm:"column:token_hash"`
	UserID         int64  `gorm:"column:user_id"`
	ExpiresAt      int64  `gorm:"column:expires_at"`
	RevokedAt      *int64 `gorm:"column:revoked_at"`
	ReplacedByHash string `gorm:"column:replaced_by_hash"`
}

type AuthTokenRepository interface {
	SaveRefreshToken(ctx context.Context, tokenHash string, userID int64, expiresAt int64) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenRecord, error)
	RotateRefreshToken(ctx context.Context, oldHash, newHash string, userID int64, newExpiresAt int64) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	BlacklistAccessToken(ctx context.Context, jti string, expiresAt int64, reason string) error
	IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

type gormAuthTokenRepository struct {
	db *gorm.DB
}

func NewGormAuthTokenRepository(db *gorm.DB) AuthTokenRepository {
	return &gormAuthTokenRepository{db: db}
}

func (r *gormAuthTokenRepository) SaveRefreshToken(ctx context.Context, tokenHash string, userID int64, expiresAt int64) error {
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Table("refresh_tokens").Create(map[string]interface{}{
		"token_hash": tokenHash,
		"user_id":    userID,
		"expires_at": expiresAt,
		"created_at": now,
		"updated_at": now,
	}).Error
}

func (r *gormAuthTokenRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenRecord, error) {
	var record RefreshTokenRecord
	err := r.db.WithContext(ctx).
		Table("refresh_tokens").
		Select("token_hash", "user_id", "expires_at", "revoked_at", "replaced_by_hash").
		Where("token_hash = ?", tokenHash).
		Limit(1).
		Take(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *gormAuthTokenRepository) RotateRefreshToken(ctx context.Context, oldHash, newHash string, userID int64, newExpiresAt int64) error {
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		update := tx.Table("refresh_tokens").
			Where("token_hash = ? AND revoked_at IS NULL", oldHash).
			Updates(map[string]interface{}{
				"revoked_at":       now,
				"replaced_by_hash": newHash,
				"updated_at":       now,
			})
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return tx.Table("refresh_tokens").Create(map[string]interface{}{
			"token_hash": newHash,
			"user_id":    userID,
			"expires_at": newExpiresAt,
			"created_at": now,
			"updated_at": now,
		}).Error
	})
}

func (r *gormAuthTokenRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	now := time.Now().Unix()
	return r.db.WithContext(ctx).
		Table("refresh_tokens").
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Updates(map[string]interface{}{
			"revoked_at": now,
			"updated_at": now,
		}).Error
}

func (r *gormAuthTokenRepository) BlacklistAccessToken(ctx context.Context, jti string, expiresAt int64, reason string) error {
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Table("token_blacklist").Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "jti"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"expires_at": expiresAt,
			"revoked_at": now,
			"reason":     reason,
		}),
	}).Create(map[string]interface{}{
		"jti":        jti,
		"expires_at": expiresAt,
		"revoked_at": now,
		"reason":     reason,
	}).Error
}

func (r *gormAuthTokenRepository) IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		var count int64
		err := r.db.WithContext(ctx).
			Table("token_blacklist").
			Where("jti = ? AND expires_at >= ?", jti, time.Now().Unix()).
			Count(&count).Error
		if err == nil {
			return count > 0, nil
		}
		lastErr = err
		if !isRetryableTokenCheckError(err) || attempt == 1 {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	return false, lastErr
}

func isRetryableTokenCheckError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid connection") ||
		strings.Contains(msg, "driver: bad connection") ||
		strings.Contains(msg, "connection reset by peer")
}
