package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProjectAnalysisPlanService interface {
	Create(ctx context.Context, input service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error)
	Update(ctx context.Context, id uint, input service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error)
	Get(ctx context.Context, id uint) (*service.ProjectAnalysisPlan, error)
	Delete(ctx context.Context, ids []uint) error
	Search(ctx context.Context, query service.ProjectAnalysisPlanSearchQuery) (*service.ProjectAnalysisPlanPage, error)
}

type ProjectAnalysisPlanHandler struct {
	plans ProjectAnalysisPlanService
}

func NewProjectAnalysisPlanHandler(plans ProjectAnalysisPlanService) *ProjectAnalysisPlanHandler {
	return &ProjectAnalysisPlanHandler{plans: plans}
}

func (h *ProjectAnalysisPlanHandler) Create(c *gin.Context) {
	var req projectAnalysisPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "分析计划参数错误")
		return
	}
	plan, err := h.plans.Create(c.Request.Context(), req.toInput())
	if err != nil {
		writeProjectAnalysisPlanError(c, err)
		return
	}
	response.Success(c, plan)
}

func (h *ProjectAnalysisPlanHandler) Update(c *gin.Context) {
	var req projectAnalysisPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "分析计划参数错误")
		return
	}
	plan, err := h.plans.Update(c.Request.Context(), req.ID, req.toInput())
	if err != nil {
		writeProjectAnalysisPlanError(c, err)
		return
	}
	response.Success(c, plan)
}

func (h *ProjectAnalysisPlanHandler) Get(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "分析计划ID不能为空")
		return
	}
	plan, err := h.plans.Get(c.Request.Context(), id)
	if err != nil {
		writeProjectAnalysisPlanError(c, err)
		return
	}
	response.Success(c, plan)
}

func (h *ProjectAnalysisPlanHandler) Delete(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择要删除的分析计划")
		return
	}
	if err := h.plans.Delete(c.Request.Context(), req.IDs); err != nil {
		writeProjectAnalysisPlanError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ProjectAnalysisPlanHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := h.plans.Search(c.Request.Context(), service.ProjectAnalysisPlanSearchQuery{
		ProjectID: parseOptionalUint(c.Query("projectId")),
		Keyword:   strings.TrimSpace(c.Query("keyword")),
		Enabled:   parseOptionalBoolPtr(c.Query("enabled")),
		Page:      page,
		Size:      size,
	})
	if err != nil {
		writeProjectAnalysisPlanError(c, err)
		return
	}
	response.Success(c, result)
}

type projectAnalysisPlanRequest struct {
	ID                uint   `json:"id"`
	ProjectID         uint   `json:"projectId"`
	Name              string `json:"name"`
	Prompt            string `json:"prompt"`
	CronExpression    string `json:"cronExpression"`
	Enabled           *bool  `json:"enabled"`
	IMEnabled         bool   `json:"imEnabled"`
	IMRobotID         uint   `json:"imRobotId"`
	HTMLReportEnabled *bool  `json:"htmlReportEnabled"`
}

func (r projectAnalysisPlanRequest) toInput() service.ProjectAnalysisPlanInput {
	return service.ProjectAnalysisPlanInput{
		ProjectID:         r.ProjectID,
		Name:              r.Name,
		Prompt:            r.Prompt,
		CronExpression:    r.CronExpression,
		Enabled:           r.Enabled,
		IMEnabled:         r.IMEnabled,
		IMRobotID:         r.IMRobotID,
		HTMLReportEnabled: r.HTMLReportEnabled,
	}
}

func writeProjectAnalysisPlanError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidProjectAnalysisPlanInput):
		response.BadRequest(c, "分析计划参数错误")
	case errors.Is(err, service.ErrProjectAnalysisPlanNotFound):
		response.Error(c, http.StatusNotFound, 40400, "分析计划不存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "分析计划操作失败")
	}
}

func parseOptionalBoolPtr(value string) *bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil
	}
	return &parsed
}
