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

func TestProjectGitLabHandlerRemoteSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gitlabSvc := &fakeProjectGitLabService{
		searchProjects: func(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabProject, error) {
			require.Equal(t, "https://gitlab.example.com", input.APIBaseURL)
			require.Equal(t, "secret", input.AccessToken)
			return []service.GitLabProject{{ID: 1, Name: "ai-review", WebURL: "https://gitlab.example.com/group/ai-review"}}, nil
		},
	}
	r := gin.New()
	r.POST("/remote-search", NewProjectGitLabHandler(gitlabSvc).RemoteSearch)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/remote-search", strings.NewReader(`{"apiBaseUrl":"https://gitlab.example.com","accessToken":"secret","keyword":"ai","page":1,"size":20}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].([]any)
	require.Len(t, data, 1)
}

func TestProjectGitLabHandlerGroupSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/group-search", NewProjectGitLabHandler(&fakeProjectGitLabService{
		searchGroups: func(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabGroup, error) {
			return []service.GitLabGroup{{ID: 1, Name: "platform", FullPath: "org/platform"}}, nil
		},
	}).GroupSearch)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/group-search", strings.NewReader(`{"apiBaseUrl":"https://gitlab.example.com","accessToken":"secret","keyword":"platform"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

type fakeProjectGitLabService struct {
	searchProjects func(context.Context, service.GitLabSearchInput) ([]service.GitLabProject, error)
	searchGroups   func(context.Context, service.GitLabSearchInput) ([]service.GitLabGroup, error)
}

func (s *fakeProjectGitLabService) SearchProjects(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabProject, error) {
	return s.searchProjects(ctx, input)
}

func (s *fakeProjectGitLabService) SearchGroups(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabGroup, error) {
	return s.searchGroups(ctx, input)
}
