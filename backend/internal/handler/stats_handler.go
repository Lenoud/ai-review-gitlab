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

type StatsService interface {
	GetStats(ctx context.Context, query service.StatsRange) (*service.StatsOverview, error)
	GetMemberCommitSummary(ctx context.Context, query service.MemberCommitSummaryQuery) (*service.MemberCommitStatsPage, error)
}

type StatsHandler struct {
	stats StatsService
}

func NewStatsHandler(stats StatsService) *StatsHandler {
	return &StatsHandler{stats: stats}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	startTime, endTime, ok := parseRequiredTimeRange(c)
	if !ok {
		return
	}
	stats, err := h.stats.GetStats(c.Request.Context(), service.StatsRange{StartTime: startTime, EndTime: endTime})
	if err != nil {
		writeStatsError(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *StatsHandler) MemberCommitSummary(c *gin.Context) {
	startTime, endTime, ok := parseRequiredTimeRange(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := h.stats.GetMemberCommitSummary(c.Request.Context(), service.MemberCommitSummaryQuery{
		StartTime: startTime,
		EndTime:   endTime,
		Project:   c.Query("project"),
		Page:      page,
		Size:      size,
	})
	if err != nil {
		writeStatsError(c, err)
		return
	}
	response.Success(c, result)
}

func parseRequiredTimeRange(c *gin.Context) (int64, int64, bool) {
	startTime, startErr := strconv.ParseInt(c.Query("startTime"), 10, 64)
	endTime, endErr := strconv.ParseInt(c.Query("endTime"), 10, 64)
	if startErr != nil || endErr != nil || startTime <= 0 || endTime <= 0 || startTime >= endTime {
		response.BadRequest(c, "时间范围参数错误")
		return 0, 0, false
	}
	return startTime, endTime, true
}

func writeStatsError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidStatsInput):
		response.BadRequest(c, "时间范围参数错误")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "统计数据查询失败")
	}
}
