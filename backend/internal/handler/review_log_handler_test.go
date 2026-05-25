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

func TestReviewLogHandlerDeleteAndGeneratePushShareToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewReviewLogHandler(&fakeReviewLogService{
		deletePush: func(ctx context.Context, id uint) error {
			require.Equal(t, uint(12), id)
			return nil
		},
		generatePushShareToken: func(ctx context.Context, id uint) (*service.ReviewLogShareToken, error) {
			require.Equal(t, uint(12), id)
			return &service.ReviewLogShareToken{ShareToken: "abc", ShareTokenExpiresAt: 123456}, nil
		},
		getShareToken: func(ctx context.Context, eventType string, eventID uint) (*service.ReviewLogShareToken, error) {
			require.Equal(t, "push", eventType)
			require.Equal(t, uint(12), eventID)
			return &service.ReviewLogShareToken{ShareToken: "abc", ShareTokenExpiresAt: 123456}, nil
		},
	})
	r.POST("/push-review-log/delete", handler.DeletePush)
	r.POST("/push-review-log/generate-share-token/:logId", handler.GeneratePushShareToken)
	r.GET("/review-log/get-share-token", handler.GetShareToken)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/push-review-log/delete", strings.NewReader(`{"id":12}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/push-review-log/generate-share-token/12", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, "abc", data["shareToken"])
	require.Equal(t, float64(123456), data["shareTokenExpiresAt"])

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/review-log/get-share-token?reviewEventType=push&reviewEventId=12", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestReviewLogHandlerAuthorsAndProjectNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewReviewLogHandler(&fakeReviewLogService{
		pushAuthors: func(ctx context.Context, query service.ReviewLogOptionQuery) ([]service.AuthorOption, error) {
			require.Equal(t, []string{"ai-review"}, query.ProjectNames)
			return []service.AuthorOption{{Value: "alice", Label: "alice（Alice）", DisplayName: "Alice"}}, nil
		},
		mergeProjectNames: func(ctx context.Context, query service.ReviewLogOptionQuery) ([]string, error) {
			require.Equal(t, []string{"bob"}, query.Authors)
			return []string{"ai-review"}, nil
		},
	})
	r.GET("/push-review-log/authors", handler.PushAuthors)
	r.GET("/merge-request-review-log/project-names", handler.MergeRequestProjectNames)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/push-review-log/authors?projectNames=ai-review", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/merge-request-review-log/project-names?authors=bob", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].([]any)
	require.Equal(t, "ai-review", data[0])
}

type fakeReviewLogService struct {
	getPush                func(context.Context, uint) (*service.PushReviewLog, error)
	searchPush             func(context.Context, service.ReviewLogSearchQuery) (*service.PushReviewLogPage, error)
	deletePush             func(context.Context, uint) error
	pushAuthors            func(context.Context, service.ReviewLogOptionQuery) ([]service.AuthorOption, error)
	pushProjectNames       func(context.Context, service.ReviewLogOptionQuery) ([]string, error)
	generatePushShareToken func(context.Context, uint) (*service.ReviewLogShareToken, error)
	getMerge               func(context.Context, uint) (*service.MergeRequestReviewLog, error)
	searchMerge            func(context.Context, service.ReviewLogSearchQuery) (*service.MergeRequestReviewLogPage, error)
	deleteMerge            func(context.Context, uint) error
	mergeAuthors           func(context.Context, service.ReviewLogOptionQuery) ([]service.AuthorOption, error)
	mergeProjectNames      func(context.Context, service.ReviewLogOptionQuery) ([]string, error)
	generateMRShareToken   func(context.Context, uint) (*service.ReviewLogShareToken, error)
	getShareToken          func(context.Context, string, uint) (*service.ReviewLogShareToken, error)
}

func (s *fakeReviewLogService) GetPush(ctx context.Context, id uint) (*service.PushReviewLog, error) {
	return s.getPush(ctx, id)
}

func (s *fakeReviewLogService) SearchPush(ctx context.Context, query service.ReviewLogSearchQuery) (*service.PushReviewLogPage, error) {
	return s.searchPush(ctx, query)
}

func (s *fakeReviewLogService) DeletePush(ctx context.Context, id uint) error {
	return s.deletePush(ctx, id)
}

func (s *fakeReviewLogService) PushAuthors(ctx context.Context, query service.ReviewLogOptionQuery) ([]service.AuthorOption, error) {
	return s.pushAuthors(ctx, query)
}

func (s *fakeReviewLogService) PushProjectNames(ctx context.Context, query service.ReviewLogOptionQuery) ([]string, error) {
	return s.pushProjectNames(ctx, query)
}

func (s *fakeReviewLogService) GeneratePushShareToken(ctx context.Context, id uint) (*service.ReviewLogShareToken, error) {
	return s.generatePushShareToken(ctx, id)
}

func (s *fakeReviewLogService) GetMergeRequest(ctx context.Context, id uint) (*service.MergeRequestReviewLog, error) {
	return s.getMerge(ctx, id)
}

func (s *fakeReviewLogService) SearchMergeRequest(ctx context.Context, query service.ReviewLogSearchQuery) (*service.MergeRequestReviewLogPage, error) {
	return s.searchMerge(ctx, query)
}

func (s *fakeReviewLogService) DeleteMergeRequest(ctx context.Context, id uint) error {
	return s.deleteMerge(ctx, id)
}

func (s *fakeReviewLogService) MergeRequestAuthors(ctx context.Context, query service.ReviewLogOptionQuery) ([]service.AuthorOption, error) {
	return s.mergeAuthors(ctx, query)
}

func (s *fakeReviewLogService) MergeRequestProjectNames(ctx context.Context, query service.ReviewLogOptionQuery) ([]string, error) {
	return s.mergeProjectNames(ctx, query)
}

func (s *fakeReviewLogService) GenerateMergeRequestShareToken(ctx context.Context, id uint) (*service.ReviewLogShareToken, error) {
	return s.generateMRShareToken(ctx, id)
}

func (s *fakeReviewLogService) GetShareToken(ctx context.Context, eventType string, eventID uint) (*service.ReviewLogShareToken, error) {
	return s.getShareToken(ctx, eventType, eventID)
}
