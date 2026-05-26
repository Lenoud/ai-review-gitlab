package handler

import (
	"context"
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type RBACService interface {
	ListRoles(ctx context.Context) ([]service.Role, error)
	ListPermissionGroups(ctx context.Context) ([]service.PermissionGroup, error)
}

type RBACHandler struct {
	rbac RBACService
}

func NewRBACHandler(rbac RBACService) *RBACHandler {
	return &RBACHandler{rbac: rbac}
}

func (h *RBACHandler) ListRoles(c *gin.Context) {
	roles, err := h.rbac.ListRoles(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "角色列表查询失败")
		return
	}
	response.Success(c, roles)
}

func (h *RBACHandler) MenuPermissions(c *gin.Context) {
	groups, err := h.rbac.ListPermissionGroups(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "权限列表查询失败")
		return
	}
	response.Success(c, groups)
}
