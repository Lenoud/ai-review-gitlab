package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestAnalysisExecutionLogHandlerTestRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAnalysisExecutionLogHandler(&fakeAnalysisExecutionLogService{
		testRun: func(ctx context.Context, input service.AnalysisTestRunInput) (*service.ProjectAnalysisPlanExecutionLog, error) {
			require.Equal(t, uint(7), input.ProjectID)
			require.Equal(t, uint(3), input.PlanID)
			require.Equal(t, "summarize", input.Prompt)
			require.Equal(t, "weekly", input.PlanName)
			require.True(t, input.IMEnabled)
			require.Equal(t, uint(5), input.IMRobotID)
			require.True(t, input.HTMLReportEnabled)
			return &service.ProjectAnalysisPlanExecutionLog{ID: 22, ProjectID: 7, PlanID: 3, Status: "succeeded", ResultContent: "analysis"}, nil
		},
	})
	r.POST("/test-run", handler.TestRun)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test-run", strings.NewReader(`{"projectId":7,"planId":3,"prompt":"summarize","planName":"weekly","imEnabled":true,"imRobotId":5,"htmlReportEnabled":true}`))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(22), data["id"])
	require.Equal(t, "succeeded", data["status"])
}

func TestAnalysisExecutionLogHandlerTestRunMapsValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewAnalysisExecutionLogHandler(&fakeAnalysisExecutionLogService{
		testRun: func(ctx context.Context, input service.AnalysisTestRunInput) (*service.ProjectAnalysisPlanExecutionLog, error) {
			return nil, service.ErrInvalidReviewLogInput
		},
	})
	r.POST("/test-run", handler.TestRun)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test-run", strings.NewReader(`{"projectId":7,"prompt":" "}`))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalysisExecutionLogHandlerTestRunMapsMissingDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		err  error
	}{
		{name: "project", err: service.ErrProjectNotFound},
		{name: "model", err: service.ErrLLMModelNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewAnalysisExecutionLogHandler(&fakeAnalysisExecutionLogService{
				testRun: func(ctx context.Context, input service.AnalysisTestRunInput) (*service.ProjectAnalysisPlanExecutionLog, error) {
					return nil, tt.err
				},
			})
			r.POST("/test-run", handler.TestRun)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/test-run", strings.NewReader(`{"projectId":7,"prompt":"summarize"}`))
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusNotFound, w.Code)
		})
	}
}

type fakeAnalysisExecutionLogService struct {
	get                func(context.Context, uint) (*service.ProjectAnalysisPlanExecutionLog, error)
	search             func(context.Context, service.AnalysisExecutionLogSearchQuery) (*service.AnalysisExecutionLogPage, error)
	generateShareToken func(context.Context, uint) (*service.ReviewLogShareToken, error)
	testRun            func(context.Context, service.AnalysisTestRunInput) (*service.ProjectAnalysisPlanExecutionLog, error)
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

func (s *fakeAnalysisExecutionLogService) TestRun(ctx context.Context, input service.AnalysisTestRunInput) (*service.ProjectAnalysisPlanExecutionLog, error) {
	return s.testRun(ctx, input)
}
