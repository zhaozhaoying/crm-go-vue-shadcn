package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidRole     = errors.New("invalid role")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidUserIDs  = errors.New("invalid user ids")
)

const defaultUserAvatar = "https://zhaozhaoying.oss-accelerate.aliyuncs.com/avatars/2026/03/18/1773818260823402723.jpg"

type CreateUserInput struct {
	Username           string
	Password           string
	Nickname           string
	Email              string
	Mobile             string
	HanghangCRMMobile  string
	Avatar             string
	RoleID             int64
	ParentID           *int64
}

type UpdateUserInput struct {
	Username           string
	Password           string
	Nickname           string
	Email              string
	Mobile             string
	HanghangCRMMobile  string
	Avatar             string
	RoleID             int64
	ParentID           *int64
	Status             string
}

type UserService interface {
	List(ctx context.Context) ([]model.UserWithRole, error)
	Search(ctx context.Context, keyword string) ([]model.UserWithRole, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Create(ctx context.Context, input CreateUserInput) (*model.User, error)
	Update(ctx context.Context, id int64, input UpdateUserInput) (*model.User, error)
	BatchDisable(ctx context.Context, ids []int64) (int64, error)
	Delete(ctx context.Context, id int64) error
}

type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) UserService {
	return &userService{userRepo: userRepo, roleRepo: roleRepo}
}

func normalizeUserAvatar(avatar string) string {
	if strings.TrimSpace(avatar) == "" {
		return defaultUserAvatar
	}
	return avatar
}

func (s *userService) List(ctx context.Context) ([]model.UserWithRole, error) {
	return s.userRepo.ListWithRole(ctx)
}

func (s *userService) Search(ctx context.Context, keyword string) ([]model.UserWithRole, error) {
	if keyword == "" {
		return s.userRepo.ListWithRole(ctx)
	}
	return s.userRepo.SearchWithRole(ctx, keyword)
}

func (s *userService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *userService) Create(ctx context.Context, input CreateUserInput) (*model.User, error) {
	// 校验角色
	if _, err := s.roleRepo.FindByID(ctx, input.RoleID); err != nil {
		return nil, ErrInvalidRole
	}
	// 校验用户名唯一
	if existing, _ := s.userRepo.FindByUsername(ctx, input.Username); existing != nil {
		return nil, ErrUserExists
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:           input.Username,
		Password:           string(hashed),
		Nickname:           input.Nickname,
		Email:              input.Email,
		Mobile:             input.Mobile,
		HanghangCRMMobile:  strings.TrimSpace(input.HanghangCRMMobile),
		Avatar:             normalizeUserAvatar(input.Avatar),
		RoleID:             input.RoleID,
		ParentID:           input.ParentID,
		Status:             model.UserStatusEnabled,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, id int64, input UpdateUserInput) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if _, err := s.roleRepo.FindByID(ctx, input.RoleID); err != nil {
		return nil, ErrInvalidRole
	}
	user.Username = input.Username
	user.Nickname = input.Nickname
	user.Email = input.Email
	user.Mobile = input.Mobile
	user.HanghangCRMMobile = strings.TrimSpace(input.HanghangCRMMobile)
	user.Avatar = normalizeUserAvatar(input.Avatar)
	user.RoleID = input.RoleID
	user.ParentID = input.ParentID
	if input.Password != "" {
		if len(input.Password) < 6 {
			return nil, ErrInvalidPassword
		}
		hashed, hashErr := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, hashErr
		}
		user.Password = string(hashed)
	}
	if input.Status == model.UserStatusEnabled || input.Status == model.UserStatusDisabled {
		user.Status = input.Status
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Delete(ctx context.Context, id int64) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *userService) BatchDisable(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, ErrInvalidUserIDs
	}

	unique := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return 0, ErrInvalidUserIDs
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	return s.userRepo.BatchUpdateStatus(ctx, unique, model.UserStatusDisabled)
}
