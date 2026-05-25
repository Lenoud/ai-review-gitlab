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

func TestReviewLogHandlerSearchPushParsesFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/push-review-log/search", NewReviewLogHandler(&fakeReviewLogService{
		searchPush: func(ctx context.Context, query service.ReviewLogSearchQuery) (*service.PushReviewLogPage, error) {
			require.Equal(t, uint(7), query.ProjectID)
			require.Equal(t, []string{"alice", "bob"}, query.Authors)
			require.Equal(t, []string{"ai-review"}, query.ProjectNames)
			require.Equal(t, "main", query.Branch)
			require.NotNil(t, query.MinScore)
			require.Equal(t, 60, *query.MinScore)
			require.Equal(t, 2, query.Page)
			require.Equal(t, 10, query.Size)
			return &service.PushReviewLogPage{
				Items: []service.PushReviewLog{{ID: 1, ProjectName: "ai-review", Branch: "main"}},
				Total: 1,
				Page:  2,
				Size:  10,
			}, nil
		},
	}).SearchPush)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/push-review-log/search?projectId=7&authors=alice,bob&projectNames=ai-review&branch=main&minScore=60&page=2&size=10", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["total"])
}

func TestReviewLogHandlerGetMergeRequestNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/merge-request-review-log/get", NewReviewLogHandler(&fakeReviewLogService{
		getMerge: func(ctx context.Context, id uint) (*service.MergeRequestReviewLog, error) {
			require.Equal(t, uint(99), id)
			return nil, service.ErrReviewLogNotFound
		},
	}).GetMergeRequest)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/merge-request-review-log/get?id=99", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

type fakeReviewLogService struct {
	getPush     func(context.Context, uint) (*service.PushReviewLog, error)
	searchPush  func(context.Context, service.ReviewLogSearchQuery) (*service.PushReviewLogPage, error)
	getMerge    func(context.Context, uint) (*service.MergeRequestReviewLog, error)
	searchMerge func(context.Context, service.ReviewLogSearchQuery) (*service.MergeRequestReviewLogPage, error)
}

func (s *fakeReviewLogService) GetPush(ctx context.Context, id uint) (*service.PushReviewLog, error) {
	return s.getPush(ctx, id)
}

func (s *fakeReviewLogService) SearchPush(ctx context.Context, query service.ReviewLogSearchQuery) (*service.PushReviewLogPage, error) {
	return s.searchPush(ctx, query)
}

func (s *fakeReviewLogService) GetMergeRequest(ctx context.Context, id uint) (*service.MergeRequestReviewLog, error) {
	return s.getMerge(ctx, id)
}

func (s *fakeReviewLogService) SearchMergeRequest(ctx context.Context, query service.ReviewLogSearchQuery) (*service.MergeRequestReviewLogPage, error) {
	return s.searchMerge(ctx, query)
}
