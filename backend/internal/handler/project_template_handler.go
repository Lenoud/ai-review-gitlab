package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProjectTemplateService interface {
	Create(ctx context.Context, input service.ProjectTemplateInput) (*service.ProjectTemplate, error)
	Update(ctx context.Context, id uint, input service.ProjectTemplateInput) (*service.ProjectTemplate, error)
	Get(ctx context.Context, id uint) (*service.ProjectTemplate, error)
	Delete(ctx context.Context, ids []uint) error
	List(ctx context.Context, query service.ProjectTemplateListQuery) ([]service.ProjectTemplate, error)
}

type ProjectTemplateHandler struct {
	templates ProjectTemplateService
}

func NewProjectTemplateHandler(templates ProjectTemplateService) *ProjectTemplateHandler {
	return &ProjectTemplateHandler{templates: templates}
}

func (h *ProjectTemplateHandler) Create(c *gin.Context) {
	var req projectTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "项目模板参数错误")
		return
	}
	template, err := h.templates.Create(c.Request.Context(), req.toInput())
	if err != nil {
		writeProjectTemplateError(c, err)
		return
	}
	response.Success(c, template)
}

func (h *ProjectTemplateHandler) Update(c *gin.Context) {
	var req projectTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "项目模板参数错误")
		return
	}
	template, err := h.templates.Update(c.Request.Context(), req.ID, req.toInput())
	if err != nil {
		writeProjectTemplateError(c, err)
		return
	}
	response.Success(c, template)
}

func (h *ProjectTemplateHandler) Get(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "项目模板ID不能为空")
		return
	}
	template, err := h.templates.Get(c.Request.Context(), id)
	if err != nil {
		writeProjectTemplateError(c, err)
		return
	}
	response.Success(c, template)
}

func (h *ProjectTemplateHandler) Delete(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择要删除的项目模板")
		return
	}
	if err := h.templates.Delete(c.Request.Context(), req.IDs); err != nil {
		writeProjectTemplateError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ProjectTemplateHandler) List(c *gin.Context) {
	result, err := h.templates.List(c.Request.Context(), service.ProjectTemplateListQuery{
		Keyword: strings.TrimSpace(c.Query("keyword")),
	})
	if err != nil {
		writeProjectTemplateError(c, err)
		return
	}
	response.Success(c, result)
}

type projectTemplateRequest struct {
	ID                   uint     `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Extensions           []string `json:"extensions"`
	ReviewPromptTemplate string   `json:"reviewPromptTemplate"`
}

func (r projectTemplateRequest) toInput() service.ProjectTemplateInput {
	return service.ProjectTemplateInput{
		Name:                 r.Name,
		Description:          r.Description,
		Extensions:           r.Extensions,
		ReviewPromptTemplate: r.ReviewPromptTemplate,
	}
}

func writeProjectTemplateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidProjectTemplateInput):
		response.BadRequest(c, "项目模板参数错误")
	case errors.Is(err, service.ErrProjectTemplateNotFound):
		response.Error(c, http.StatusNotFound, 40400, "项目模板不存在")
	case errors.Is(err, service.ErrProjectTemplateInUse):
		response.Error(c, http.StatusConflict, 40900, "项目模板已被项目使用")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "项目模板操作失败")
	}
}
