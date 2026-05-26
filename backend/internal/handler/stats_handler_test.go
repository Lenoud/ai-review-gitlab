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

func TestStatsHandlerGetStatsReturnsOverview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/stats", NewStatsHandler(&fakeStatsService{
		overview: service.StatsOverview{ActiveProjects: 2, Contributors: 3, TotalCommits: 4, AverageScore: 88.5},
	}).GetStats)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/stats?startTime=1000&endTime=2000", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(2), data["activeProjects"])
	require.Equal(t, float64(4), data["totalCommits"])
}

func TestStatsHandlerGetStatsRejectsInvalidRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/stats", NewStatsHandler(&fakeStatsService{}).GetStats)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/stats?startTime=2000&endTime=1000", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStatsHandlerMemberCommitSummaryReturnsPagedStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/member/commit-summary", NewStatsHandler(&fakeStatsService{
		memberPage: service.MemberCommitStatsPage{
			Items: []service.MemberCommitStats{{
				AuthorIdentity:    "alice",
				AuthorDisplayText: "alice（Alice）",
				Summary:           service.MemberCommitSummary{CommitCount: 2},
			}},
			Total: 1,
			Page:  1,
			Size:  20,
		},
	}).MemberCommitSummary)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/member/commit-summary?startTime=1000&endTime=2000&project=api", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["total"])
	items := data["items"].([]any)
	require.Equal(t, "alice（Alice）", items[0].(map[string]any)["authorDisplayText"])
}

type fakeStatsService struct {
	overview   service.StatsOverview
	memberPage service.MemberCommitStatsPage
	err        error
}

func (s *fakeStatsService) GetStats(ctx context.Context, query service.StatsRange) (*service.StatsOverview, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.overview, nil
}

func (s *fakeStatsService) GetMemberCommitSummary(ctx context.Context, query service.MemberCommitSummaryQuery) (*service.MemberCommitStatsPage, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.memberPage, nil
}
