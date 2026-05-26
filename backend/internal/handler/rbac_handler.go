package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

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
	CreateUser(ctx context.Context, input service.AdminUserInput) (*service.AdminUser, error)
	UpdateUser(ctx context.Context, id uint, input service.AdminUserInput) (*service.AdminUser, error)
	GetUser(ctx context.Context, id uint) (*service.AdminUser, error)
	SearchUsers(ctx context.Context, query service.AdminUserSearchQuery) (*service.AdminUserPage, error)
	ListRoleOptions(ctx context.Context) ([]service.Role, error)
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

func (h *RBACHandler) CreateUser(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "用户参数错误")
		return
	}
	user, err := h.rbac.CreateUser(c.Request.Context(), req.toInput())
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, user)
}

func (h *RBACHandler) UpdateUser(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "用户参数错误")
		return
	}
	user, err := h.rbac.UpdateUser(c.Request.Context(), req.ID, req.toInput())
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, user)
}

func (h *RBACHandler) GetUser(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "用户ID不能为空")
		return
	}
	user, err := h.rbac.GetUser(c.Request.Context(), id)
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, user)
}

func (h *RBACHandler) SearchUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := h.rbac.SearchUsers(c.Request.Context(), service.AdminUserSearchQuery{
		Keyword: c.Query("keyword"),
		Page:    page,
		Size:    size,
	})
	if err != nil {
		writeUserError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *RBACHandler) RoleOptions(c *gin.Context) {
	roles, err := h.rbac.ListRoleOptions(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50000, "角色选项查询失败")
		return
	}
	response.Success(c, roles)
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

type userRequest struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Remark   string `json:"remark"`
	RoleIDs  []uint `json:"roleIds"`
}

func (r userRequest) toInput() service.AdminUserInput {
	return service.AdminUserInput{
		Username: r.Username,
		Password: r.Password,
		Nickname: r.Nickname,
		Remark:   r.Remark,
		RoleIDs:  r.RoleIDs,
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

func writeUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidRBACInput):
		response.BadRequest(c, "用户参数错误")
	case errors.Is(err, service.ErrUserNotFound):
		response.Error(c, http.StatusNotFound, 40400, "用户不存在")
	case errors.Is(err, service.ErrUsernameExists):
		response.Error(c, http.StatusConflict, 40900, "用户名已存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "用户操作失败")
	}
}
