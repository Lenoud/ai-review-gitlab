package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ReviewLogService interface {
	GetPush(ctx context.Context, id uint) (*service.PushReviewLog, error)
	SearchPush(ctx context.Context, query service.ReviewLogSearchQuery) (*service.PushReviewLogPage, error)
	GetMergeRequest(ctx context.Context, id uint) (*service.MergeRequestReviewLog, error)
	SearchMergeRequest(ctx context.Context, query service.ReviewLogSearchQuery) (*service.MergeRequestReviewLogPage, error)
}

type ReviewLogHandler struct {
	logs ReviewLogService
}

func NewReviewLogHandler(logs ReviewLogService) *ReviewLogHandler {
	return &ReviewLogHandler{logs: logs}
}

func (h *ReviewLogHandler) GetPush(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "审查日志ID不能为空")
		return
	}
	log, err := h.logs.GetPush(c.Request.Context(), id)
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, log)
}

func (h *ReviewLogHandler) SearchPush(c *gin.Context) {
	page, err := h.logs.SearchPush(c.Request.Context(), parseReviewLogSearchQuery(c))
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, page)
}

func (h *ReviewLogHandler) GetMergeRequest(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "审查日志ID不能为空")
		return
	}
	log, err := h.logs.GetMergeRequest(c.Request.Context(), id)
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, log)
}

func (h *ReviewLogHandler) SearchMergeRequest(c *gin.Context) {
	page, err := h.logs.SearchMergeRequest(c.Request.Context(), parseReviewLogSearchQuery(c))
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, page)
}

func parseReviewLogSearchQuery(c *gin.Context) service.ReviewLogSearchQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	query := service.ReviewLogSearchQuery{
		ProjectID:      parseOptionalUint(c.Query("projectId")),
		Authors:        parseListQuery(c, "authors"),
		ProjectNames:   parseListQuery(c, "projectNames"),
		Branch:         c.Query("branch"),
		CommitMessages: c.Query("commitMessages"),
		MinScore:       parseOptionalIntPtr(c.Query("minScore")),
		MaxScore:       parseOptionalIntPtr(c.Query("maxScore")),
		StartTime:      parseOptionalUnixMilliPtr(c.Query("startTime")),
		EndTime:        parseOptionalUnixMilliPtr(c.Query("endTime")),
		Page:           page,
		Size:           size,
	}
	return query
}

func parseOptionalUint(value string) uint {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return uint(parsed)
}

func parseOptionalIntPtr(value string) *int {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseOptionalUnixMilliPtr(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	t := time.UnixMilli(parsed)
	return &t
}

func parseListQuery(c *gin.Context, key string) []string {
	values := c.QueryArray(key)
	if len(values) == 0 {
		values = []string{c.Query(key)}
	}
	items := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				items = append(items, part)
			}
		}
	}
	return items
}

func writeReviewLogError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidReviewLogInput):
		response.BadRequest(c, "审查日志参数错误")
	case errors.Is(err, service.ErrReviewLogNotFound):
		response.Error(c, http.StatusNotFound, 40400, "审查日志不存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "审查日志操作失败")
	}
}
