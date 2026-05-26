package service

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidRBACInput = errors.New("invalid rbac input")
	ErrRoleNotFound     = errors.New("role not found")
	ErrRoleCodeExists   = errors.New("role code exists")
	ErrRoleInUse        = errors.New("role in use")
)

type Role struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type RoleDetail struct {
	ID            uint   `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permissionIds"`
}

type RoleInput struct {
	Code          string
	Name          string
	Description   string
	PermissionIDs []uint
}

type Permission struct {
	ID       uint   `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type PermissionGroup struct {
	Category    string       `json:"category"`
	Permissions []Permission `json:"permissions"`
}

type RBACRepository interface {
	ListRoles(ctx context.Context) ([]Role, error)
	ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error)
	CreateRole(ctx context.Context, input RoleInput) (*RoleDetail, error)
	UpdateRole(ctx context.Context, id uint, input RoleInput) (*RoleDetail, error)
	FindRoleByID(ctx context.Context, id uint) (*RoleDetail, error)
	DeleteRoles(ctx context.Context, ids []uint) error
}

type RBACService struct {
	rbac RBACRepository
}

func NewRBACService(rbac RBACRepository) *RBACService {
	return &RBACService{rbac: rbac}
}

func (s *RBACService) ListRoles(ctx context.Context) ([]Role, error) {
	return s.rbac.ListRoles(ctx)
}

func (s *RBACService) ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error) {
	return s.rbac.ListPermissionGroups(ctx)
}

func (s *RBACService) CreateRole(ctx context.Context, input RoleInput) (*RoleDetail, error) {
	normalized, err := normalizeRoleInput(input)
	if err != nil {
		return nil, err
	}
	return s.rbac.CreateRole(ctx, normalized)
}

func (s *RBACService) UpdateRole(ctx context.Context, id uint, input RoleInput) (*RoleDetail, error) {
	if id == 0 {
		return nil, ErrInvalidRBACInput
	}
	normalized, err := normalizeRoleInput(input)
	if err != nil {
		return nil, err
	}
	return s.rbac.UpdateRole(ctx, id, normalized)
}

func (s *RBACService) GetRole(ctx context.Context, id uint) (*RoleDetail, error) {
	if id == 0 {
		return nil, ErrInvalidRBACInput
	}
	return s.rbac.FindRoleByID(ctx, id)
}

func (s *RBACService) DeleteRoles(ctx context.Context, ids []uint) error {
	cleanIDs := cleanUintIDs(ids)
	if len(cleanIDs) == 0 {
		return ErrInvalidRBACInput
	}
	return s.rbac.DeleteRoles(ctx, cleanIDs)
}

func normalizeRoleInput(input RoleInput) (RoleInput, error) {
	input.Code = strings.TrimSpace(input.Code)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.PermissionIDs = cleanUintIDs(input.PermissionIDs)
	if input.Code == "" || input.Name == "" {
		return RoleInput{}, ErrInvalidRBACInput
	}
	if len(input.Code) > 64 || len(input.Name) > 64 || len(input.Description) > 255 {
		return RoleInput{}, ErrInvalidRBACInput
	}
	return input, nil
}

func cleanUintIDs(ids []uint) []uint {
	clean := make([]uint, 0, len(ids))
	seen := map[uint]struct{}{}
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		clean = append(clean, id)
	}
	return clean
}
