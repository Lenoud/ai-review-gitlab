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

type MemberIMMappingService interface {
	Create(ctx context.Context, input service.MemberIMMappingInput) (*service.MemberIMMapping, error)
	Update(ctx context.Context, id uint, input service.MemberIMMappingInput) (*service.MemberIMMapping, error)
	Get(ctx context.Context, id uint) (*service.MemberIMMapping, error)
	Delete(ctx context.Context, ids []uint) error
	Search(ctx context.Context, query service.MemberIMMappingSearchQuery) (*service.MemberIMMappingPage, error)
}

type MemberIMMappingHandler struct {
	mappings MemberIMMappingService
}

func NewMemberIMMappingHandler(mappings MemberIMMappingService) *MemberIMMappingHandler {
	return &MemberIMMappingHandler{mappings: mappings}
}

func (h *MemberIMMappingHandler) Create(c *gin.Context) {
	var req memberIMMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "成员IM映射参数错误")
		return
	}
	mapping, err := h.mappings.Create(c.Request.Context(), req.toInput())
	if err != nil {
		writeMemberIMMappingError(c, err)
		return
	}
	response.Success(c, mapping)
}

func (h *MemberIMMappingHandler) Update(c *gin.Context) {
	var req memberIMMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "成员IM映射参数错误")
		return
	}
	mapping, err := h.mappings.Update(c.Request.Context(), req.ID, req.toInput())
	if err != nil {
		writeMemberIMMappingError(c, err)
		return
	}
	response.Success(c, mapping)
}

func (h *MemberIMMappingHandler) Get(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "成员IM映射ID不能为空")
		return
	}
	mapping, err := h.mappings.Get(c.Request.Context(), id)
	if err != nil {
		writeMemberIMMappingError(c, err)
		return
	}
	response.Success(c, mapping)
}

func (h *MemberIMMappingHandler) Delete(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择要删除的成员IM映射")
		return
	}
	if err := h.mappings.Delete(c.Request.Context(), req.IDs); err != nil {
		writeMemberIMMappingError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *MemberIMMappingHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := h.mappings.Search(c.Request.Context(), service.MemberIMMappingSearchQuery{
		Keyword:  c.Query("keyword"),
		Platform: c.Query("platform"),
		Page:     page,
		Size:     size,
	})
	if err != nil {
		writeMemberIMMappingError(c, err)
		return
	}
	response.Success(c, result)
}

type memberIMMappingRequest struct {
	ID          uint   `json:"id"`
	GitUsername string `json:"gitUsername"`
	Platform    string `json:"platform"`
	IMUserID    string `json:"imUserId"`
	DisplayName string `json:"displayName"`
}

func (r memberIMMappingRequest) toInput() service.MemberIMMappingInput {
	return service.MemberIMMappingInput{
		GitUsername: r.GitUsername,
		Platform:    r.Platform,
		IMUserID:    r.IMUserID,
		DisplayName: r.DisplayName,
	}
}

func writeMemberIMMappingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidMemberIMMappingInput):
		response.BadRequest(c, "成员IM映射参数错误")
	case errors.Is(err, service.ErrMemberIMMappingNotFound):
		response.Error(c, http.StatusNotFound, 40400, "成员IM映射不存在")
	case errors.Is(err, service.ErrMemberIMMappingExists):
		response.Error(c, http.StatusConflict, 40900, "成员IM映射已存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "成员IM映射操作失败")
	}
}
