package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProjectTemplateReviewRuleService interface {
	Create(ctx context.Context, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error)
	Update(ctx context.Context, id uint, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error)
	Get(ctx context.Context, id uint) (*service.ProjectTemplateReviewRule, error)
	Delete(ctx context.Context, id uint) error
	ListByTemplateID(ctx context.Context, templateID uint) ([]service.ProjectTemplateReviewRule, error)
}

type ProjectTemplateReviewRuleHandler struct {
	rules ProjectTemplateReviewRuleService
}

func NewProjectTemplateReviewRuleHandler(rules ProjectTemplateReviewRuleService) *ProjectTemplateReviewRuleHandler {
	return &ProjectTemplateReviewRuleHandler{rules: rules}
}

func (h *ProjectTemplateReviewRuleHandler) Create(c *gin.Context) {
	var req projectTemplateReviewRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "项目模板审查规则参数错误")
		return
	}
	rule, err := h.rules.Create(c.Request.Context(), req.toInput())
	if err != nil {
		writeProjectTemplateReviewRuleError(c, err)
		return
	}
	response.Success(c, rule)
}

func (h *ProjectTemplateReviewRuleHandler) Update(c *gin.Context) {
	var req projectTemplateReviewRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "项目模板审查规则参数错误")
		return
	}
	rule, err := h.rules.Update(c.Request.Context(), req.ID, req.toInput())
	if err != nil {
		writeProjectTemplateReviewRuleError(c, err)
		return
	}
	response.Success(c, rule)
}

func (h *ProjectTemplateReviewRuleHandler) Get(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "审查规则ID不能为空")
		return
	}
	rule, err := h.rules.Get(c.Request.Context(), id)
	if err != nil {
		writeProjectTemplateReviewRuleError(c, err)
		return
	}
	response.Success(c, rule)
}

func (h *ProjectTemplateReviewRuleHandler) Delete(c *gin.Context) {
	var req idRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "审查规则ID不能为空")
		return
	}
	if err := h.rules.Delete(c.Request.Context(), req.ID); err != nil {
		writeProjectTemplateReviewRuleError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ProjectTemplateReviewRuleHandler) ListByTemplateID(c *gin.Context) {
	templateID, ok := parseUintQuery(c, "projectTemplateId")
	if !ok {
		templateID, ok = parseUintQuery(c, "templateId")
	}
	if !ok {
		response.BadRequest(c, "项目模板ID不能为空")
		return
	}
	rules, err := h.rules.ListByTemplateID(c.Request.Context(), templateID)
	if err != nil {
		writeProjectTemplateReviewRuleError(c, err)
		return
	}
	response.Success(c, rules)
}

type projectTemplateReviewRuleRequest struct {
	ID                uint     `json:"id"`
	ProjectTemplateID uint     `json:"projectTemplateId"`
	TemplateID        uint     `json:"templateId"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	GlobPatterns      []string `json:"globPatterns"`
	Content           string   `json:"content"`
	Priority          int      `json:"priority"`
	Enabled           *bool    `json:"enabled"`
}

func (r projectTemplateReviewRuleRequest) toInput() service.ProjectTemplateReviewRuleInput {
	templateID := r.ProjectTemplateID
	if templateID == 0 {
		templateID = r.TemplateID
	}
	input := service.ProjectTemplateReviewRuleInput{
		TemplateID:   templateID,
		Name:         r.Name,
		Description:  r.Description,
		GlobPatterns: r.GlobPatterns,
		Content:      r.Content,
		Priority:     r.Priority,
	}
	if r.Enabled != nil {
		input.Enabled = *r.Enabled
		input.EnabledIsSet = true
	}
	return input
}

func writeProjectTemplateReviewRuleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidProjectTemplateReviewRuleInput):
		response.BadRequest(c, "项目模板审查规则参数错误")
	case errors.Is(err, service.ErrProjectTemplateNotFound):
		response.Error(c, http.StatusNotFound, 40400, "项目模板不存在")
	case errors.Is(err, service.ErrProjectTemplateReviewRuleNotFound):
		response.Error(c, http.StatusNotFound, 40400, "审查规则不存在")
	case errors.Is(err, service.ErrProjectTemplateReviewRuleTemplateMismatch):
		response.BadRequest(c, "审查规则不属于指定项目模板")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "项目模板审查规则操作失败")
	}
}
