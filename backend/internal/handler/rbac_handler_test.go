package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

type fakeRBACService struct {
	roles            []service.Role
	permissionGroups []service.PermissionGroup
}

func (s *fakeRBACService) ListRoles(ctx context.Context) ([]service.Role, error) {
	return s.roles, nil
}

func (s *fakeRBACService) ListPermissionGroups(ctx context.Context) ([]service.PermissionGroup, error) {
	return s.permissionGroups, nil
}
