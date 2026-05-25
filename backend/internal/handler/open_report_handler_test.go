package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenReportHandlerCodeReviewReportReturnsHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/code-review-report", NewOpenReportHandler(&fakeOpenReportService{
		codeReviewReport: func(ctx context.Context, input service.CodeReviewReportInput) (string, error) {
			require.Equal(t, uint(12), input.LogID)
			require.Equal(t, "push", input.LogType)
			require.Equal(t, "token", input.Token)
			return "<!doctype html><html><body><h1>AI Review</h1></body></html>", nil
		},
	}).CodeReviewReport)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/code-review-report?logId=12&logType=push&token=token", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.True(t, strings.HasPrefix(w.Body.String(), "<!doctype html>"))
}

func TestOpenReportHandlerCodeReviewReportNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/code-review-report", NewOpenReportHandler(&fakeOpenReportService{
		codeReviewReport: func(ctx context.Context, input service.CodeReviewReportInput) (string, error) {
			return "", service.ErrReviewLogNotFound
		},
	}).CodeReviewReport)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/code-review-report?logId=12&logType=push&token=bad", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

type fakeOpenReportService struct {
	codeReviewReport func(context.Context, service.CodeReviewReportInput) (string, error)
}

func (s *fakeOpenReportService) CodeReviewReport(ctx context.Context, input service.CodeReviewReportInput) (string, error) {
	return s.codeReviewReport(ctx, input)
}
