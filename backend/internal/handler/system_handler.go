package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SystemService interface {
	GetConfig(ctx context.Context) (*service.SystemConfig, error)
	UpdateBaseURL(ctx context.Context, baseURL string) (*service.SystemConfig, error)
}

type SystemHandler struct {
	system SystemService
}

func NewSystemHandler(system SystemService) *SystemHandler {
	return &SystemHandler{system: system}
}

func (h *SystemHandler) OpenInfo(c *gin.Context) {
	config, err := h.system.GetConfig(c.Request.Context())
	if err != nil {
		writeSystemError(c, err)
		return
	}
	response.Success(c, gin.H{
		"siteName":   config.SiteName,
		"siteNotice": config.SiteNotice,
	})
}

func (h *SystemHandler) Info(c *gin.Context) {
	config, err := h.system.GetConfig(c.Request.Context())
	if err != nil {
		writeSystemError(c, err)
		return
	}
	response.Success(c, config)
}

func (h *SystemHandler) Config(c *gin.Context) {
	config, err := h.system.GetConfig(c.Request.Context())
	if err != nil {
		writeSystemError(c, err)
		return
	}
	response.Success(c, config)
}

func (h *SystemHandler) UpdateBaseURL(c *gin.Context) {
	var req baseURLUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "系统配置参数错误")
		return
	}
	config, err := h.system.UpdateBaseURL(c.Request.Context(), req.BaseURL)
	if err != nil {
		writeSystemError(c, err)
		return
	}
	response.Success(c, config)
}

type baseURLUpdateRequest struct {
	BaseURL string `json:"baseUrl"`
}

func writeSystemError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidSystemConfigInput):
		response.BadRequest(c, "系统配置参数错误")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "系统配置操作失败")
	}
}
