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

type AnalysisExecutionLogService interface {
	Get(ctx context.Context, id uint) (*service.ProjectAnalysisPlanExecutionLog, error)
	Search(ctx context.Context, query service.AnalysisExecutionLogSearchQuery) (*service.AnalysisExecutionLogPage, error)
	GenerateShareToken(ctx context.Context, id uint) (*service.ReviewLogShareToken, error)
}

type AnalysisExecutionLogHandler struct {
	logs AnalysisExecutionLogService
}

func NewAnalysisExecutionLogHandler(logs AnalysisExecutionLogService) *AnalysisExecutionLogHandler {
	return &AnalysisExecutionLogHandler{logs: logs}
}

func (h *AnalysisExecutionLogHandler) Get(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "分析执行日志ID不能为空")
		return
	}
	log, err := h.logs.Get(c.Request.Context(), id)
	if err != nil {
		writeAnalysisExecutionLogError(c, err)
		return
	}
	response.Success(c, log)
}

func (h *AnalysisExecutionLogHandler) Search(c *gin.Context) {
	page, err := h.logs.Search(c.Request.Context(), parseAnalysisExecutionLogSearchQuery(c))
	if err != nil {
		writeAnalysisExecutionLogError(c, err)
		return
	}
	response.Success(c, page)
}

func (h *AnalysisExecutionLogHandler) GenerateShareToken(c *gin.Context) {
	id, ok := parseUintParam(c, "logId")
	if !ok {
		response.BadRequest(c, "分析执行日志ID不能为空")
		return
	}
	token, err := h.logs.GenerateShareToken(c.Request.Context(), id)
	if err != nil {
		writeAnalysisExecutionLogError(c, err)
		return
	}
	response.Success(c, token)
}

func (h *AnalysisExecutionLogHandler) HTMLReport(c *gin.Context) {
	id, ok := parseUintParam(c, "logId")
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}
	log, err := h.logs.Get(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidReviewLogInput):
			c.Status(http.StatusBadRequest)
		case errors.Is(err, service.ErrReviewLogNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(service.BuildAnalysisExecutionReportHTML(log)))
}

func parseAnalysisExecutionLogSearchQuery(c *gin.Context) service.AnalysisExecutionLogSearchQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	return service.AnalysisExecutionLogSearchQuery{
		ProjectID: parseOptionalUint(c.Query("projectId")),
		PlanID:    parseOptionalUint(c.Query("planId")),
		Status:    strings.TrimSpace(c.Query("status")),
		StartTime: parseOptionalUnixMilliPtr(firstNonBlankQuery(c, "startTimestamp", "startTime")),
		EndTime:   parseOptionalUnixMilliPtr(firstNonBlankQuery(c, "endTimestamp", "endTime")),
		Page:      page,
		Size:      size,
	}
}

func writeAnalysisExecutionLogError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidReviewLogInput):
		response.BadRequest(c, "分析执行日志参数错误")
	case errors.Is(err, service.ErrReviewLogNotFound):
		response.Error(c, http.StatusNotFound, 40400, "分析执行日志不存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "分析执行日志操作失败")
	}
}
