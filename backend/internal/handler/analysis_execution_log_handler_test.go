package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAnalysisExecutionLogHandlerGetAndSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAnalysisExecutionLogHandler(&fakeAnalysisExecutionLogService{
		get: func(ctx context.Context, id uint) (*service.ProjectAnalysisPlanExecutionLog, error) {
			require.Equal(t, uint(12), id)
			return &service.ProjectAnalysisPlanExecutionLog{ID: 12, ProjectID: 7, PlanID: 3, Status: "succeeded"}, nil
		},
		search: func(ctx context.Context, query service.AnalysisExecutionLogSearchQuery) (*service.AnalysisExecutionLogPage, error) {
			require.Equal(t, uint(7), query.ProjectID)
			require.Equal(t, uint(3), query.PlanID)
			require.Equal(t, "succeeded", query.Status)
			require.Equal(t, 2, query.Page)
			require.Equal(t, 10, query.Size)
			return &service.AnalysisExecutionLogPage{
				Items: []service.ProjectAnalysisPlanExecutionLog{{ID: 12, Status: "succeeded"}},
				Total: 1,
				Page:  2,
				Size:  10,
			}, nil
		},
	})
	r.GET("/get", handler.Get)
	r.GET("/search", handler.Search)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/get?id=12", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/search?projectId=7&planId=3&status=succeeded&page=2&size=10", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["total"])
}

func TestAnalysisExecutionLogHandlerGenerateShareTokenAndHTMLReport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAnalysisExecutionLogHandler(&fakeAnalysisExecutionLogService{
		get: func(ctx context.Context, id uint) (*service.ProjectAnalysisPlanExecutionLog, error) {
			require.Equal(t, uint(12), id)
			return &service.ProjectAnalysisPlanExecutionLog{
				ID:            12,
				ProjectID:     7,
				PlanID:        3,
				Status:        "succeeded",
				ResultContent: "analysis content",
				DurationMs:    1200,
			}, nil
		},
		generateShareToken: func(ctx context.Context, id uint) (*service.ReviewLogShareToken, error) {
			require.Equal(t, uint(12), id)
			return &service.ReviewLogShareToken{ShareToken: "abc", ShareTokenExpiresAt: 123456}, nil
		},
	})
	r.POST("/generate-share-token/:logId", handler.GenerateShareToken)
	r.GET("/html-report/:logId", handler.HTMLReport)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/generate-share-token/12", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, "abc", data["shareToken"])

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/html-report/12", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Body.String(), "analysis content")
}

type fakeAnalysisExecutionLogService struct {
	get                func(context.Context, uint) (*service.ProjectAnalysisPlanExecutionLog, error)
	search             func(context.Context, service.AnalysisExecutionLogSearchQuery) (*service.AnalysisExecutionLogPage, error)
	generateShareToken func(context.Context, uint) (*service.ReviewLogShareToken, error)
}

func (s *fakeAnalysisExecutionLogService) Get(ctx context.Context, id uint) (*service.ProjectAnalysisPlanExecutionLog, error) {
	return s.get(ctx, id)
}

func (s *fakeAnalysisExecutionLogService) Search(ctx context.Context, query service.AnalysisExecutionLogSearchQuery) (*service.AnalysisExecutionLogPage, error) {
	return s.search(ctx, query)
}

func (s *fakeAnalysisExecutionLogService) GenerateShareToken(ctx context.Context, id uint) (*service.ReviewLogShareToken, error) {
	return s.generateShareToken(ctx, id)
}
