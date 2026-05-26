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

type SysLogService interface {
	Get(ctx context.Context, id uint) (*service.SysLog, error)
	Search(ctx context.Context, query service.SysLogSearchQuery) (*service.SysLogPage, error)
}

type SysLogHandler struct {
	logs SysLogService
}

func NewSysLogHandler(logs SysLogService) *SysLogHandler {
	return &SysLogHandler{logs: logs}
}

func (h *SysLogHandler) Get(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "系统日志ID不能为空")
		return
	}
	log, err := h.logs.Get(c.Request.Context(), id)
	if err != nil {
		writeSysLogError(c, err)
		return
	}
	response.Success(c, log)
}

func (h *SysLogHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	startTime, endTime, ok := parseOptionalTimeRange(c)
	if !ok {
		return
	}
	result, err := h.logs.Search(c.Request.Context(), service.SysLogSearchQuery{
		Level:     c.Query("level"),
		Module:    c.Query("module"),
		Action:    c.Query("action"),
		Message:   c.Query("message"),
		StartTime: startTime,
		EndTime:   endTime,
		Page:      page,
		Size:      size,
	})
	if err != nil {
		writeSysLogError(c, err)
		return
	}
	response.Success(c, result)
}

func parseOptionalTimeRange(c *gin.Context) (*time.Time, *time.Time, bool) {
	startTime, startOK := parseOptionalUnixMilliQuery(c, "startTime")
	endTime, endOK := parseOptionalUnixMilliQuery(c, "endTime")
	if !startOK || !endOK {
		response.BadRequest(c, "系统日志时间参数错误")
		return nil, nil, false
	}
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		response.BadRequest(c, "系统日志时间参数错误")
		return nil, nil, false
	}
	return startTime, endTime, true
}

func parseOptionalUnixMilliQuery(c *gin.Context, key string) (*time.Time, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, true
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return nil, false
	}
	parsed := time.UnixMilli(value)
	return &parsed, true
}

func writeSysLogError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidSysLogInput):
		response.BadRequest(c, "系统日志参数错误")
	case errors.Is(err, service.ErrSysLogNotFound):
		response.Error(c, http.StatusNotFound, 40400, "系统日志不存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "系统日志查询失败")
	}
}
