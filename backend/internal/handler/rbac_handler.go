package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type RBACService interface {
	ListRoles(ctx context.Context) ([]service.Role, error)
	ListPermissionGroups(ctx context.Context) ([]service.PermissionGroup, error)
	CreateRole(ctx context.Context, input service.RoleInput) (*service.RoleDetail, error)
	UpdateRole(ctx context.Context, id uint, input service.RoleInput) (*service.RoleDetail, error)
	GetRole(ctx context.Context, id uint) (*service.RoleDetail, error)
	DeleteRoles(ctx context.Context, ids []uint) error
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

func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "角色参数错误")
		return
	}
	role, err := h.rbac.CreateRole(c.Request.Context(), req.toInput())
	if err != nil {
		writeRoleError(c, err)
		return
	}
	response.Success(c, role)
}

func (h *RBACHandler) UpdateRole(c *gin.Context) {
	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "角色参数错误")
		return
	}
	role, err := h.rbac.UpdateRole(c.Request.Context(), req.ID, req.toInput())
	if err != nil {
		writeRoleError(c, err)
		return
	}
	response.Success(c, role)
}

func (h *RBACHandler) GetRole(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "角色ID不能为空")
		return
	}
	role, err := h.rbac.GetRole(c.Request.Context(), id)
	if err != nil {
		writeRoleError(c, err)
		return
	}
	response.Success(c, role)
}

func (h *RBACHandler) DeleteRole(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择要删除的角色")
		return
	}
	if err := h.rbac.DeleteRoles(c.Request.Context(), req.IDs); err != nil {
		writeRoleError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

type roleRequest struct {
	ID            uint   `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permissionIds"`
}

func (r roleRequest) toInput() service.RoleInput {
	return service.RoleInput{
		Code:          r.Code,
		Name:          r.Name,
		Description:   r.Description,
		PermissionIDs: r.PermissionIDs,
	}
}

func writeRoleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidRBACInput):
		response.BadRequest(c, "角色参数错误")
	case errors.Is(err, service.ErrRoleNotFound):
		response.Error(c, http.StatusNotFound, 40400, "角色不存在")
	case errors.Is(err, service.ErrRoleCodeExists):
		response.Error(c, http.StatusConflict, 40900, "角色编码已存在")
	case errors.Is(err, service.ErrRoleInUse):
		response.Error(c, http.StatusConflict, 40900, "角色已被用户使用")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "角色操作失败")
	}
}
