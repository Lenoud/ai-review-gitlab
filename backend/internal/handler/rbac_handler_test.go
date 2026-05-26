package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRBACHandlerListRolesReturnsRoleOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/roles", NewRBACHandler(&fakeRBACService{
		roles: []service.Role{
			{ID: 1, Code: "admin", Name: "管理员"},
		},
	}).ListRoles)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].([]any)
	role := data[0].(map[string]any)
	require.Equal(t, float64(1), role["id"])
	require.Equal(t, "admin", role["code"])
	require.Equal(t, "管理员", role["name"])
}

func TestRBACHandlerMenuPermissionsReturnsPermissionGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/permissions", NewRBACHandler(&fakeRBACService{
		permissionGroups: []service.PermissionGroup{
			{
				Category: "project",
				Permissions: []service.Permission{
					{ID: 1, Code: "project:read", Name: "查看项目", Category: "project"},
				},
			},
		},
	}).MenuPermissions)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].([]any)
	group := data[0].(map[string]any)
	require.Equal(t, "project", group["category"])
	permissions := group["permissions"].([]any)
	permission := permissions[0].(map[string]any)
	require.Equal(t, "project:read", permission["code"])
}

func TestRBACHandlerCreateRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeRBACService{
		roleDetail: &service.RoleDetail{ID: 3, Code: "reviewer", Name: "审查员", PermissionIDs: []uint{1, 2}},
	}
	r := gin.New()
	r.POST("/roles", NewRBACHandler(fake).CreateRole)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{"code":"reviewer","name":"审查员","permissionIds":[1,2]}`))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "reviewer", fake.created.Code)
	require.Equal(t, []uint{1, 2}, fake.created.PermissionIDs)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, "reviewer", data["code"])
}

func TestRBACHandlerUpdateRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeRBACService{
		roleDetail: &service.RoleDetail{ID: 3, Code: "maintainer", Name: "维护者"},
	}
	r := gin.New()
	r.POST("/roles", NewRBACHandler(fake).UpdateRole)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{"id":3,"code":"maintainer","name":"维护者"}`))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, uint(3), fake.updatedID)
	require.Equal(t, "maintainer", fake.updated.Code)
}

func TestRBACHandlerGetRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/roles", NewRBACHandler(&fakeRBACService{
		roleDetail: &service.RoleDetail{ID: 3, Code: "reviewer", Name: "审查员", PermissionIDs: []uint{1}},
	}).GetRole)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/roles?id=3", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(3), data["id"])
}

func TestRBACHandlerDeleteRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeRBACService{}
	r := gin.New()
	r.POST("/roles/delete", NewRBACHandler(fake).DeleteRole)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/roles/delete", strings.NewReader(`{"ids":[3,4]}`))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, []uint{3, 4}, fake.deletedIDs)
}

func TestRBACHandlerRoleConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/roles", NewRBACHandler(&fakeRBACService{err: service.ErrRoleCodeExists}).CreateRole)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{"code":"admin","name":"管理员"}`))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

type fakeRBACService struct {
	roles            []service.Role
	permissionGroups []service.PermissionGroup
	roleDetail       *service.RoleDetail
	created          service.RoleInput
	updatedID        uint
	updated          service.RoleInput
	deletedIDs       []uint
	err              error
}

func (s *fakeRBACService) ListRoles(ctx context.Context) ([]service.Role, error) {
	return s.roles, nil
}

func (s *fakeRBACService) ListPermissionGroups(ctx context.Context) ([]service.PermissionGroup, error) {
	return s.permissionGroups, nil
}

func (s *fakeRBACService) CreateRole(ctx context.Context, input service.RoleInput) (*service.RoleDetail, error) {
	s.created = input
	if s.err != nil {
		return nil, s.err
	}
	return s.roleDetail, nil
}

func (s *fakeRBACService) UpdateRole(ctx context.Context, id uint, input service.RoleInput) (*service.RoleDetail, error) {
	s.updatedID = id
	s.updated = input
	if s.err != nil {
		return nil, s.err
	}
	return s.roleDetail, nil
}

func (s *fakeRBACService) GetRole(ctx context.Context, id uint) (*service.RoleDetail, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.roleDetail, nil
}

func (s *fakeRBACService) DeleteRoles(ctx context.Context, ids []uint) error {
	s.deletedIDs = ids
	return s.err
}
