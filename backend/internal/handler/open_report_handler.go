package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type OpenReportService interface {
	CodeReviewReport(ctx context.Context, input service.CodeReviewReportInput) (string, error)
}

type OpenReportHandler struct {
	reports OpenReportService
}

func NewOpenReportHandler(reports OpenReportService) *OpenReportHandler {
	return &OpenReportHandler{reports: reports}
}

func (h *OpenReportHandler) CodeReviewReport(c *gin.Context) {
	logID, err := strconv.ParseUint(strings.TrimSpace(c.Query("logId")), 10, 64)
	if err != nil || logID == 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	html, err := h.reports.CodeReviewReport(c.Request.Context(), service.CodeReviewReportInput{
		LogID:   uint(logID),
		LogType: c.Query("logType"),
		Token:   c.Query("token"),
	})
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
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
