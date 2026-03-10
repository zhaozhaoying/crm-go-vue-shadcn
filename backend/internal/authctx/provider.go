package authctx

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrUserNotFound = errors.New("user not found")
)

type Claims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenJTI string `json:"tokenJti"`
	TokenExp int64  `json:"tokenExp"`
}

type CurrentUser struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	Mobile    string    `json:"mobile"`
	Avatar    string    `json:"avatar"`
	RoleID    int64     `json:"roleId"`
	RoleName  string    `json:"roleName"`
	RoleLabel string    `json:"roleLabel"`
	ParentID  *int64    `json:"parentId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Provider interface {
	GetClaims(c *gin.Context) (*Claims, error)
	GetCurrentUser(ctx context.Context, c *gin.Context) (*CurrentUser, error)
}

type provider struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewProvider(userRepo repository.UserRepository, roleRepo repository.RoleRepository) Provider {
	return &provider{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (p *provider) GetClaims(c *gin.Context) (*Claims, error) {
	return GetClaimsFromContext(c)
}

func GetClaimsFromContext(c *gin.Context) (*Claims, error) {
	userIDValue, ok := c.Get("userID")
	if !ok {
		return nil, ErrUnauthorized
	}
	userID, ok := toInt64(userIDValue)
	if !ok {
		return nil, ErrUnauthorized
	}

	username, _ := c.Get("username")
	role, _ := c.Get("role")
	tokenJTI, _ := c.Get("tokenJTI")
	tokenExp, _ := c.Get("tokenExp")

	usernameStr, _ := username.(string)
	roleStr, _ := role.(string)
	tokenJTIStr, _ := tokenJTI.(string)
	tokenExpInt, _ := toInt64(tokenExp)

	return &Claims{
		UserID:   userID,
		Username: usernameStr,
		Role:     roleStr,
		TokenJTI: tokenJTIStr,
		TokenExp: tokenExpInt,
	}, nil
}

func GetUserIDFromContext(c *gin.Context) (int64, error) {
	claims, err := GetClaimsFromContext(c)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

func (p *provider) GetCurrentUser(ctx context.Context, c *gin.Context) (*CurrentUser, error) {
	claims, err := p.GetClaims(c)
	if err != nil {
		return nil, err
	}

	user, err := p.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	roleName := ""
	roleLabel := ""
	if role, err := p.roleRepo.FindByID(ctx, user.RoleID); err == nil && role != nil {
		roleName = role.Name
		roleLabel = role.Label
	}

	return mapUser(user, roleName, roleLabel), nil
}

func mapUser(user *model.User, roleName, roleLabel string) *CurrentUser {
	return &CurrentUser{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Mobile:    user.Mobile,
		Avatar:    user.Avatar,
		RoleID:    user.RoleID,
		RoleName:  roleName,
		RoleLabel: roleLabel,
		ParentID:  user.ParentID,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func toInt64(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
