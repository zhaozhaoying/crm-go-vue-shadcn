package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
)

var ErrRoleNotFound = errors.New("role not found")

type RoleService interface {
	List(ctx context.Context) ([]model.Role, error)
	Create(ctx context.Context, name, label string, sort int) (*model.Role, error)
	Update(ctx context.Context, id int64, name, label string, sort int) (*model.Role, error)
	Delete(ctx context.Context, id int64) error
}

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) List(ctx context.Context) ([]model.Role, error) {
	return s.repo.List(ctx)
}

func (s *roleService) Create(ctx context.Context, name, label string, sort int) (*model.Role, error) {
	role := &model.Role{Name: name, Label: label, Sort: sort}
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Update(ctx context.Context, id int64, name, label string, sort int) (*model.Role, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrRoleNotFound
	}
	role.Name = name
	role.Label = label
	role.Sort = sort
	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
