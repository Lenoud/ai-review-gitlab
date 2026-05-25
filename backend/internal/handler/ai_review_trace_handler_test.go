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

func TestAIReviewTraceHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/ai-review-trace/create", NewAIReviewTraceHandler(&fakeAIReviewTraceService{
		create: func(ctx context.Context, input service.AIReviewTraceInput) (*service.AIReviewTrace, error) {
			require.Equal(t, "push", input.ReviewEventType)
			require.Equal(t, uint(101), input.ReviewEventID)
			require.Equal(t, "prompt", input.Prompt)
			require.Equal(t, "response", input.Response)
			require.Equal(t, "openai", input.Provider)
			require.Equal(t, "gpt-test", input.ModelCode)
			return &service.AIReviewTrace{ID: 1, ReviewEventType: "push", ReviewEventID: 101}, nil
		},
	}).Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ai-review-trace/create", strings.NewReader(`{
		"reviewEventType":"push",
		"reviewEventId":101,
		"prompt":"prompt",
		"response":"response",
		"provider":"openai",
		"modelCode":"gpt-test"
	}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["id"])
}

func TestAIReviewTraceHandlerGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ai-review-trace/get", NewAIReviewTraceHandler(&fakeAIReviewTraceService{
		get: func(ctx context.Context, eventType string, eventID uint) (*service.AIReviewTrace, error) {
			require.Equal(t, "merge_request", eventType)
			require.Equal(t, uint(202), eventID)
			return &service.AIReviewTrace{ID: 2, ReviewEventType: "merge_request", ReviewEventID: 202, Response: "ok"}, nil
		},
	}).Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai-review-trace/get?reviewEventType=merge_request&reviewEventId=202", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, "ok", data["response"])
}

func TestAIReviewTraceHandlerGetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ai-review-trace/get", NewAIReviewTraceHandler(&fakeAIReviewTraceService{
		get: func(ctx context.Context, eventType string, eventID uint) (*service.AIReviewTrace, error) {
			return nil, service.ErrAIReviewTraceNotFound
		},
	}).Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ai-review-trace/get?reviewEventType=push&reviewEventId=101", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

type fakeAIReviewTraceService struct {
	create func(context.Context, service.AIReviewTraceInput) (*service.AIReviewTrace, error)
	get    func(context.Context, string, uint) (*service.AIReviewTrace, error)
}

func (s *fakeAIReviewTraceService) Create(ctx context.Context, input service.AIReviewTraceInput) (*service.AIReviewTrace, error) {
	return s.create(ctx, input)
}

func (s *fakeAIReviewTraceService) Get(ctx context.Context, eventType string, eventID uint) (*service.AIReviewTrace, error) {
	return s.get(ctx, eventType, eventID)
}
