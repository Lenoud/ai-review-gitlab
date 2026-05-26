package service

import "context"

type Role struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
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
