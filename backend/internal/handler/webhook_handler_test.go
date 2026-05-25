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

func TestWebhookHandlerEnqueuesGitLabPush(t *testing.T) {
	gin.SetMode(gin.TestMode)
	webhookSvc := &fakeReviewTaskService{
		enqueue: func(ctx context.Context, input service.GitLabWebhookInput) (*service.ReviewTaskEnqueueResult, error) {
			require.Equal(t, "Push Hook", input.Event)
			require.Contains(t, string(input.Payload), `"after":"abc123"`)
			return &service.ReviewTaskEnqueueResult{TaskID: 1, Status: service.ReviewTaskStatusPending}, nil
		},
	}
	r := gin.New()
	r.POST("/review/webhook", NewWebhookHandler(webhookSvc).GitLab)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/review/webhook", strings.NewReader(`{"after":"abc123"}`))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["taskId"])
}

func TestWebhookHandlerRejectsUnknownProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/review/webhook", NewWebhookHandler(&fakeReviewTaskService{
		enqueue: func(ctx context.Context, input service.GitLabWebhookInput) (*service.ReviewTaskEnqueueResult, error) {
			return nil, service.ErrReviewTaskProjectNotFound
		},
	}).GitLab)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/review/webhook", strings.NewReader(`{}`))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

type fakeReviewTaskService struct {
	enqueue func(context.Context, service.GitLabWebhookInput) (*service.ReviewTaskEnqueueResult, error)
}

func (s *fakeReviewTaskService) EnqueueGitLabWebhook(ctx context.Context, input service.GitLabWebhookInput) (*service.ReviewTaskEnqueueResult, error) {
	return s.enqueue(ctx, input)
}
