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
	ErrUsernameExists   = errors.New("username exists")
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

type AdminUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Remark   string `json:"remark"`
	RoleIDs  []uint `json:"roleIds"`
	Roles    []Role `json:"roles"`
}

type AdminUserInput struct {
	Username     string
	Password     string
	PasswordHash string
	Nickname     string
	Remark       string
	RoleIDs      []uint
}

type AdminUserSearchQuery struct {
	Keyword string
	Page    int
	Size    int
}

type AdminUserPage struct {
	Items []AdminUser `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

type RBACRepository interface {
	ListRoles(ctx context.Context) ([]Role, error)
	ListPermissionGroups(ctx context.Context) ([]PermissionGroup, error)
	CreateRole(ctx context.Context, input RoleInput) (*RoleDetail, error)
	UpdateRole(ctx context.Context, id uint, input RoleInput) (*RoleDetail, error)
	FindRoleByID(ctx context.Context, id uint) (*RoleDetail, error)
	DeleteRoles(ctx context.Context, ids []uint) error
	CreateUser(ctx context.Context, input AdminUserInput) (*AdminUser, error)
	UpdateUser(ctx context.Context, id uint, input AdminUserInput) (*AdminUser, error)
	FindAdminUserByID(ctx context.Context, id uint) (*AdminUser, error)
	SearchUsers(ctx context.Context, query AdminUserSearchQuery) (*AdminUserPage, error)
	ListRoleOptions(ctx context.Context) ([]Role, error)
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

func (s *RBACService) CreateUser(ctx context.Context, input AdminUserInput) (*AdminUser, error) {
	normalized, err := normalizeAdminUserInput(input, true)
	if err != nil {
		return nil, err
	}
	hash, err := HashPassword(normalized.Password)
	if err != nil {
		return nil, err
	}
	normalized.PasswordHash = hash
	normalized.Password = ""
	return s.rbac.CreateUser(ctx, normalized)
}

func (s *RBACService) UpdateUser(ctx context.Context, id uint, input AdminUserInput) (*AdminUser, error) {
	if id == 0 {
		return nil, ErrInvalidRBACInput
	}
	normalized, err := normalizeAdminUserInput(input, false)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(normalized.Password) != "" {
		hash, err := HashPassword(normalized.Password)
		if err != nil {
			return nil, err
		}
		normalized.PasswordHash = hash
	}
	normalized.Password = ""
	return s.rbac.UpdateUser(ctx, id, normalized)
}

func (s *RBACService) GetUser(ctx context.Context, id uint) (*AdminUser, error) {
	if id == 0 {
		return nil, ErrInvalidRBACInput
	}
	return s.rbac.FindAdminUserByID(ctx, id)
}

func (s *RBACService) SearchUsers(ctx context.Context, query AdminUserSearchQuery) (*AdminUserPage, error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 20
	}
	if query.Size > 200 {
		query.Size = 200
	}
	return s.rbac.SearchUsers(ctx, query)
}

func (s *RBACService) ListRoleOptions(ctx context.Context) ([]Role, error) {
	return s.rbac.ListRoleOptions(ctx)
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

func normalizeAdminUserInput(input AdminUserInput, requireUsernameAndPassword bool) (AdminUserInput, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Password = strings.TrimSpace(input.Password)
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.Remark = strings.TrimSpace(input.Remark)
	input.RoleIDs = cleanUintIDs(input.RoleIDs)
	if len(input.RoleIDs) == 0 {
		return AdminUserInput{}, ErrInvalidRBACInput
	}
	if requireUsernameAndPassword && (input.Username == "" || input.Password == "") {
		return AdminUserInput{}, ErrInvalidRBACInput
	}
	if input.Password != "" && (len(input.Password) < 6 || len(input.Password) > 64) {
		return AdminUserInput{}, ErrInvalidRBACInput
	}
	if len(input.Username) > 64 || len(input.Nickname) > 64 || len(input.Remark) > 255 {
		return AdminUserInput{}, ErrInvalidRBACInput
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
