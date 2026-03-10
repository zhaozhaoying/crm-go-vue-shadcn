package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists          = errors.New("username already exists")
	ErrInvalidCredential   = errors.New("invalid username or password")
	ErrUserDisabled        = errors.New("user is disabled")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrTokenRevoked        = errors.New("token revoked")
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (*LoginTokens, error)
	Refresh(ctx context.Context, refreshToken string) (*LoginTokens, error)
	Logout(ctx context.Context, accessJTI string, accessExp int64, refreshToken string) error
}

type authService struct {
	userRepo           repository.UserRepository
	roleRepo           repository.RoleRepository
	authTokenRepo      repository.AuthTokenRepository
	jwtSecret          []byte
	jwtExpiry          time.Duration
	refreshTokenExpiry time.Duration
}

type LoginTokens struct {
	Token            string `json:"token"`
	RefreshToken     string `json:"refreshToken"`
	ExpiresInSeconds int64  `json:"expiresInSeconds"`
}

func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	authTokenRepo repository.AuthTokenRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) AuthService {
	return &authService{
		userRepo:           userRepo,
		roleRepo:           roleRepo,
		authTokenRepo:      authTokenRepo,
		jwtSecret:          []byte(jwtSecret),
		jwtExpiry:          jwtExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (*LoginTokens, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredential
		}
		return nil, err
	}

	if user.Status == model.UserStatusDisabled {
		return nil, ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredential
	}

	roleName, err := s.resolveRoleName(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	accessToken, _, accessExp, err := s.issueToken(user, roleName, "access", s.jwtExpiry)
	if err != nil {
		return nil, err
	}
	refreshToken, _, _, err := s.issueToken(user, roleName, "refresh", s.refreshTokenExpiry)
	if err != nil {
		return nil, err
	}

	refreshTokenHash := hashToken(refreshToken)
	if err := s.authTokenRepo.SaveRefreshToken(ctx, refreshTokenHash, user.ID, time.Now().Add(s.refreshTokenExpiry).Unix()); err != nil {
		return nil, err
	}

	return &LoginTokens{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		ExpiresInSeconds: accessExp - time.Now().Unix(),
	}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*LoginTokens, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	claims, err := s.parseAndValidateToken(refreshToken, "refresh")
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	userID, err := extractUserIDFromClaims(claims)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	oldHash := hashToken(refreshToken)
	record, err := s.authTokenRepo.GetRefreshToken(ctx, oldHash)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrInvalidRefreshToken
	}
	if record.RevokedAt != nil || record.ExpiresAt < time.Now().Unix() {
		return nil, ErrTokenRevoked
	}
	if record.UserID != userID {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}
	if user.Status == model.UserStatusDisabled {
		return nil, ErrUserDisabled
	}

	roleName, err := s.resolveRoleName(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	accessToken, _, accessExp, err := s.issueToken(user, roleName, "access", s.jwtExpiry)
	if err != nil {
		return nil, err
	}
	newRefreshToken, _, _, err := s.issueToken(user, roleName, "refresh", s.refreshTokenExpiry)
	if err != nil {
		return nil, err
	}

	newHash := hashToken(newRefreshToken)
	if err := s.authTokenRepo.RotateRefreshToken(
		ctx,
		oldHash,
		newHash,
		user.ID,
		time.Now().Add(s.refreshTokenExpiry).Unix(),
	); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenRevoked
		}
		return nil, err
	}

	return &LoginTokens{
		Token:            accessToken,
		RefreshToken:     newRefreshToken,
		ExpiresInSeconds: accessExp - time.Now().Unix(),
	}, nil
}

func (s *authService) Logout(ctx context.Context, accessJTI string, accessExp int64, refreshToken string) error {
	now := time.Now().Unix()
	accessJTI = strings.TrimSpace(accessJTI)
	if accessJTI != "" && accessExp > now {
		if err := s.authTokenRepo.BlacklistAccessToken(ctx, accessJTI, accessExp, "logout"); err != nil {
			return err
		}
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil
	}
	tokenHash := hashToken(refreshToken)
	return s.authTokenRepo.RevokeRefreshToken(ctx, tokenHash)
}

func (s *authService) resolveRoleName(ctx context.Context, roleID int64) (string, error) {
	if role, err := s.roleRepo.FindByID(ctx, roleID); err == nil {
		return role.Name, nil
	}
	return "", nil
}

func (s *authService) issueToken(user *model.User, roleName, tokenType string, ttl time.Duration) (string, string, int64, error) {
	jti, err := generateTokenID()
	if err != nil {
		return "", "", 0, err
	}
	exp := time.Now().Add(ttl).Unix()
	claims := jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"role":     roleName,
		"role_id":  user.RoleID,
		"typ":      tokenType,
		"jti":      jti,
		"exp":      exp,
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", 0, err
	}
	return signed, jti, exp, nil
}

func (s *authService) parseAndValidateToken(rawToken string, expectedType string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(rawToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidRefreshToken
	}
	tokenType, _ := claims["typ"].(string)
	if tokenType != expectedType {
		return nil, ErrInvalidRefreshToken
	}
	return claims, nil
}

func extractUserIDFromClaims(claims jwt.MapClaims) (int64, error) {
	switch sub := claims["sub"].(type) {
	case float64:
		return int64(sub), nil
	case int64:
		return sub, nil
	case int:
		return int64(sub), nil
	case string:
		parsed, err := strconv.ParseInt(sub, 10, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("invalid sub claim type: %T", claims["sub"])
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generateTokenID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
