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
	DeletePush(ctx context.Context, id uint) error
	PushAuthors(ctx context.Context, query service.ReviewLogOptionQuery) ([]service.AuthorOption, error)
	PushProjectNames(ctx context.Context, query service.ReviewLogOptionQuery) ([]string, error)
	GeneratePushShareToken(ctx context.Context, id uint) (*service.ReviewLogShareToken, error)
	GetMergeRequest(ctx context.Context, id uint) (*service.MergeRequestReviewLog, error)
	SearchMergeRequest(ctx context.Context, query service.ReviewLogSearchQuery) (*service.MergeRequestReviewLogPage, error)
	DeleteMergeRequest(ctx context.Context, id uint) error
	MergeRequestAuthors(ctx context.Context, query service.ReviewLogOptionQuery) ([]service.AuthorOption, error)
	MergeRequestProjectNames(ctx context.Context, query service.ReviewLogOptionQuery) ([]string, error)
	GenerateMergeRequestShareToken(ctx context.Context, id uint) (*service.ReviewLogShareToken, error)
	GetShareToken(ctx context.Context, eventType string, eventID uint) (*service.ReviewLogShareToken, error)
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

func (h *ReviewLogHandler) DeletePush(c *gin.Context) {
	id, ok := bindReviewLogID(c)
	if !ok {
		return
	}
	if err := h.logs.DeletePush(c.Request.Context(), id); err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ReviewLogHandler) PushAuthors(c *gin.Context) {
	authors, err := h.logs.PushAuthors(c.Request.Context(), parseReviewLogOptionQuery(c))
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, authors)
}

func (h *ReviewLogHandler) PushProjectNames(c *gin.Context) {
	names, err := h.logs.PushProjectNames(c.Request.Context(), parseReviewLogOptionQuery(c))
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, names)
}

func (h *ReviewLogHandler) GeneratePushShareToken(c *gin.Context) {
	id, ok := parseUintParam(c, "logId")
	if !ok {
		response.BadRequest(c, "审查日志ID不能为空")
		return
	}
	token, err := h.logs.GeneratePushShareToken(c.Request.Context(), id)
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, token)
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

func (h *ReviewLogHandler) DeleteMergeRequest(c *gin.Context) {
	id, ok := bindReviewLogID(c)
	if !ok {
		return
	}
	if err := h.logs.DeleteMergeRequest(c.Request.Context(), id); err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ReviewLogHandler) MergeRequestAuthors(c *gin.Context) {
	authors, err := h.logs.MergeRequestAuthors(c.Request.Context(), parseReviewLogOptionQuery(c))
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, authors)
}

func (h *ReviewLogHandler) MergeRequestProjectNames(c *gin.Context) {
	names, err := h.logs.MergeRequestProjectNames(c.Request.Context(), parseReviewLogOptionQuery(c))
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, names)
}

func (h *ReviewLogHandler) GenerateMergeRequestShareToken(c *gin.Context) {
	id, ok := parseUintParam(c, "logId")
	if !ok {
		response.BadRequest(c, "审查日志ID不能为空")
		return
	}
	token, err := h.logs.GenerateMergeRequestShareToken(c.Request.Context(), id)
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, token)
}

func (h *ReviewLogHandler) GetShareToken(c *gin.Context) {
	eventID, ok := parseUintQuery(c, "reviewEventId")
	if !ok {
		response.BadRequest(c, "审查日志ID不能为空")
		return
	}
	token, err := h.logs.GetShareToken(c.Request.Context(), c.Query("reviewEventType"), eventID)
	if err != nil {
		writeReviewLogError(c, err)
		return
	}
	response.Success(c, token)
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

func parseReviewLogOptionQuery(c *gin.Context) service.ReviewLogOptionQuery {
	return service.ReviewLogOptionQuery{
		Authors:      parseListQuery(c, "authors"),
		ProjectNames: parseListQuery(c, "projectNames"),
		StartTime:    parseOptionalUnixMilliPtr(firstNonBlankQuery(c, "startTimestamp", "startTime")),
		EndTime:      parseOptionalUnixMilliPtr(firstNonBlankQuery(c, "endTimestamp", "endTime")),
	}
}

func bindReviewLogID(c *gin.Context) (uint, bool) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		response.BadRequest(c, "审查日志ID不能为空")
		return 0, false
	}
	return req.ID, true
}

func parseUintParam(c *gin.Context, key string) (uint, bool) {
	value := strings.TrimSpace(c.Param(key))
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, false
	}
	return uint(parsed), true
}

func firstNonBlankQuery(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(c.Query(key))
		if value != "" {
			return value
		}
	}
	return ""
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
