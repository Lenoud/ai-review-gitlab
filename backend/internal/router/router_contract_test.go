package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/handler"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/middleware"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPublicRoutesAreRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newContractRouter()

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/v1/open/auth/login", `{}`},
		{http.MethodPost, "/api/v1/open/auth/refresh", `{}`},
		{http.MethodPost, "/api/v1/open/auth/logout", `{}`},
		{http.MethodPost, "/review/webhook", `{}`},
		{http.MethodGet, "/api/v1/open/system/info", ``},
		{http.MethodGet, "/api/v1/open/code-review-report", ``},
		{http.MethodGet, "/api/v1/open/analysis-report", ``},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		r.ServeHTTP(w, req)
		require.NotEqual(t, http.StatusNotFound, w.Code, "%s %s", tt.method, tt.path)
	}
}

func TestAdminRoutesRequireAuthAndReturnNotImplementedWithDevToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newContractRouter()

	routes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/v1/admin/project/create", `{}`},
		{http.MethodPost, "/api/v1/admin/project/batch-create", `{}`},
		{http.MethodPost, "/api/v1/admin/project/update", `{}`},
		{http.MethodGet, "/api/v1/admin/project/get", ``},
		{http.MethodPost, "/api/v1/admin/project/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/project/search", ``},
		{http.MethodPost, "/api/v1/admin/project/gitlab/remote-search", `{}`},
		{http.MethodPost, "/api/v1/admin/project/gitlab/group-search", `{}`},
		{http.MethodPost, "/api/v1/admin/project/web-urls/exists", `{}`},
		{http.MethodGet, "/api/v1/admin/project/review-prompt/get", ``},
		{http.MethodGet, "/api/v1/admin/project/review-prompt/default", ``},
		{http.MethodPost, "/api/v1/admin/project/review-prompt/update", `{}`},
		{http.MethodPost, "/api/v1/admin/project/review-prompt/delete", `{}`},
		{http.MethodPost, "/api/v1/admin/project/review-prompt/test", `{}`},
		{http.MethodGet, "/api/v1/admin/push-review-log/get", ``},
		{http.MethodGet, "/api/v1/admin/push-review-log/search", ``},
		{http.MethodPost, "/api/v1/admin/push-review-log/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/push-review-log/authors", ``},
		{http.MethodGet, "/api/v1/admin/push-review-log/project-names", ``},
		{http.MethodPost, "/api/v1/admin/push-review-log/generate-share-token/1", `{}`},
		{http.MethodGet, "/api/v1/admin/merge-request-review-log/get", ``},
		{http.MethodGet, "/api/v1/admin/merge-request-review-log/search", ``},
		{http.MethodPost, "/api/v1/admin/merge-request-review-log/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/merge-request-review-log/authors", ``},
		{http.MethodGet, "/api/v1/admin/merge-request-review-log/project-names", ``},
		{http.MethodPost, "/api/v1/admin/merge-request-review-log/generate-share-token/1", `{}`},
		{http.MethodPost, "/api/v1/admin/ai-review-trace/create", `{}`},
		{http.MethodGet, "/api/v1/admin/ai-review-trace/get", ``},
		{http.MethodPost, "/api/v1/admin/llm-model/create", `{}`},
		{http.MethodPost, "/api/v1/admin/llm-model/update", `{}`},
		{http.MethodGet, "/api/v1/admin/llm-model/get", ``},
		{http.MethodPost, "/api/v1/admin/llm-model/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/llm-model/search", ``},
		{http.MethodGet, "/api/v1/admin/llm-model/default", ``},
		{http.MethodPost, "/api/v1/admin/llm-model/set-default", `{}`},
		{http.MethodPost, "/api/v1/admin/llm-test/connection", `{}`},
		{http.MethodPost, "/api/v1/admin/im-robot/create", `{}`},
		{http.MethodPost, "/api/v1/admin/im-robot/update", `{}`},
		{http.MethodGet, "/api/v1/admin/im-robot/get", ``},
		{http.MethodPost, "/api/v1/admin/im-robot/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/im-robot/search", ``},
		{http.MethodGet, "/api/v1/admin/im-robot/list-enabled", ``},
		{http.MethodPost, "/api/v1/admin/im-robot/test-webhook", `{}`},
		{http.MethodPost, "/api/v1/admin/member-im-mapping/create", `{}`},
		{http.MethodPost, "/api/v1/admin/member-im-mapping/update", `{}`},
		{http.MethodGet, "/api/v1/admin/member-im-mapping/get", ``},
		{http.MethodPost, "/api/v1/admin/member-im-mapping/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/member-im-mapping/search", ``},
		{http.MethodPost, "/api/v1/admin/project-template/create", `{}`},
		{http.MethodPost, "/api/v1/admin/project-template/update", `{}`},
		{http.MethodGet, "/api/v1/admin/project-template/get", ``},
		{http.MethodGet, "/api/v1/admin/project-template/list", ``},
		{http.MethodPost, "/api/v1/admin/project-template/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/project-template/review-rule/list-by-template-id", ``},
		{http.MethodPost, "/api/v1/admin/project-template/review-rule/create", `{}`},
		{http.MethodPost, "/api/v1/admin/project-template/review-rule/update", `{}`},
		{http.MethodGet, "/api/v1/admin/project-template/review-rule/get", ``},
		{http.MethodPost, "/api/v1/admin/project-template/review-rule/delete", `{}`},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan/create", `{}`},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan/update", `{}`},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan/get", ``},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan/search", ``},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan-execution-log/test-run", `{}`},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan-execution-log/get", ``},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan-execution-log/search", ``},
		{http.MethodGet, "/api/v1/admin/project-analysis-plan-execution-log/html-report/1", ``},
		{http.MethodPost, "/api/v1/admin/project-analysis-plan-execution-log/generate-share-token/1", `{}`},
		{http.MethodPost, "/api/v1/admin/user/create", `{}`},
		{http.MethodPost, "/api/v1/admin/user/update", `{}`},
		{http.MethodGet, "/api/v1/admin/user/get", ``},
		{http.MethodGet, "/api/v1/admin/user/search", ``},
		{http.MethodGet, "/api/v1/admin/user/role-options", ``},
		{http.MethodGet, "/api/v1/admin/role/list", ``},
		{http.MethodPost, "/api/v1/admin/role/create", `{}`},
		{http.MethodPost, "/api/v1/admin/role/update", `{}`},
		{http.MethodGet, "/api/v1/admin/role/get", ``},
		{http.MethodPost, "/api/v1/admin/role/delete", `{}`},
		{http.MethodGet, "/api/v1/admin/role/menu-permissions", ``},
		{http.MethodGet, "/api/v1/admin/auth/me", ``},
		{http.MethodGet, "/api/v1/admin/stats", ``},
		{http.MethodGet, "/api/v1/admin/member/commit-summary", ``},
		{http.MethodGet, "/api/v1/admin/sys-log/get", ``},
		{http.MethodGet, "/api/v1/admin/sys-log/search", ``},
		{http.MethodGet, "/api/v1/admin/system/info", ``},
		{http.MethodGet, "/api/v1/admin/system/config", ``},
		{http.MethodPost, "/api/v1/admin/system/config/base-url", `{}`},
		{http.MethodGet, "/api/v1/admin/review-log/get-share-token", ``},
	}

	for _, route := range routes {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code, "%s %s without token", route.method, route.path)

		w = httptest.NewRecorder()
		req = httptest.NewRequest(route.method, route.path, strings.NewReader(route.body))
		req.Header.Set("Authorization", "Bearer access-token")
		r.ServeHTTP(w, req)
		expectedStatus := http.StatusNotImplemented
		switch route.path {
		case "/api/v1/admin/auth/me":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/project/create",
			"/api/v1/admin/project/batch-create",
			"/api/v1/admin/project/update",
			"/api/v1/admin/project/get",
			"/api/v1/admin/project/delete",
			"/api/v1/admin/project/search",
			"/api/v1/admin/project/gitlab/remote-search",
			"/api/v1/admin/project/gitlab/group-search",
			"/api/v1/admin/project/web-urls/exists":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/llm-model/create",
			"/api/v1/admin/llm-model/update",
			"/api/v1/admin/llm-model/get",
			"/api/v1/admin/llm-model/delete",
			"/api/v1/admin/llm-model/search",
			"/api/v1/admin/llm-model/default",
			"/api/v1/admin/llm-model/set-default",
			"/api/v1/admin/llm-test/connection":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/push-review-log/get",
			"/api/v1/admin/merge-request-review-log/get",
			"/api/v1/admin/review-log/get-share-token":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/push-review-log/search",
			"/api/v1/admin/push-review-log/authors",
			"/api/v1/admin/push-review-log/project-names",
			"/api/v1/admin/merge-request-review-log/search",
			"/api/v1/admin/merge-request-review-log/authors",
			"/api/v1/admin/merge-request-review-log/project-names":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/push-review-log/delete",
			"/api/v1/admin/merge-request-review-log/delete":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/push-review-log/generate-share-token/1",
			"/api/v1/admin/merge-request-review-log/generate-share-token/1":
			expectedStatus = http.StatusNotFound
		}
		require.Equal(t, expectedStatus, w.Code, "%s %s with token", route.method, route.path)
	}
}

func newContractRouter() *gin.Engine {
	return New(Dependencies{
		AuthHandler:          NewAuthHandlerForTest(),
		ProjectHandler:       handler.NewProjectHandler(&contractProjectService{}),
		ProjectGitLabHandler: handler.NewProjectGitLabHandler(&contractProjectGitLabService{}),
		LLMModelHandler:      handler.NewLLMModelHandler(&contractLLMModelService{}),
		ReviewLogHandler:     handler.NewReviewLogHandler(&contractReviewLogService{}),
		AuthMiddleware: middleware.JWTAuth(&contractTokenValidator{
			subject: &service.AuthSubject{
				UserID:   1,
				Username: "admin",
				Nickname: "Administrator",
			},
		}),
	})
}

func NewAuthHandlerForTest() *handler.AuthHandler {
	return handler.NewAuthHandler(&contractAuthService{})
}

type contractAuthService struct{}

func (s *contractAuthService) Login(ctx context.Context, input service.LoginInput) (*service.TokenPair, error) {
	return &service.TokenPair{
		AccessToken:      "access-token",
		RefreshToken:     "refresh-token",
		TokenType:        "Bearer",
		ExpiresIn:        1800,
		RefreshExpiresIn: 2592000,
	}, nil
}

func (s *contractAuthService) Refresh(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
	return &service.TokenPair{
		AccessToken:      "new-access-token",
		RefreshToken:     "new-refresh-token",
		TokenType:        "Bearer",
		ExpiresIn:        1800,
		RefreshExpiresIn: 2592000,
	}, nil
}

type contractTokenValidator struct {
	subject *service.AuthSubject
}

func (v *contractTokenValidator) ValidateAccessToken(ctx context.Context, token string) (*service.AuthSubject, error) {
	if strings.TrimSpace(token) == "" {
		return nil, service.ErrInvalidToken
	}
	return v.subject, nil
}

type contractProjectService struct{}

func (s *contractProjectService) Create(ctx context.Context, input service.ProjectInput) (*service.Project, error) {
	return nil, service.ErrInvalidProjectInput
}

func (s *contractProjectService) BatchCreate(ctx context.Context, inputs []service.ProjectInput) ([]service.Project, error) {
	return nil, service.ErrInvalidProjectInput
}

func (s *contractProjectService) Update(ctx context.Context, id uint, input service.ProjectInput) (*service.Project, error) {
	return nil, service.ErrInvalidProjectInput
}

func (s *contractProjectService) Get(ctx context.Context, id uint) (*service.Project, error) {
	return nil, service.ErrInvalidProjectInput
}

func (s *contractProjectService) Delete(ctx context.Context, ids []uint) error {
	return service.ErrInvalidProjectInput
}

func (s *contractProjectService) Search(ctx context.Context, query service.ProjectSearchQuery) (*service.ProjectPage, error) {
	return nil, service.ErrInvalidProjectInput
}

func (s *contractProjectService) WebURLExists(ctx context.Context, webURL string, excludeID uint) (bool, error) {
	return false, service.ErrInvalidProjectInput
}

type contractProjectGitLabService struct{}

func (s *contractProjectGitLabService) SearchProjects(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabProject, error) {
	return nil, service.ErrInvalidGitLabInput
}

func (s *contractProjectGitLabService) SearchGroups(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabGroup, error) {
	return nil, service.ErrInvalidGitLabInput
}

type contractLLMModelService struct{}

func (s *contractLLMModelService) Create(ctx context.Context, input service.LLMModelInput) (*service.LLMModel, error) {
	return nil, service.ErrInvalidLLMModelInput
}

func (s *contractLLMModelService) Update(ctx context.Context, id uint, input service.LLMModelInput) (*service.LLMModel, error) {
	return nil, service.ErrInvalidLLMModelInput
}

func (s *contractLLMModelService) Get(ctx context.Context, id uint) (*service.LLMModel, error) {
	return nil, service.ErrInvalidLLMModelInput
}

func (s *contractLLMModelService) Delete(ctx context.Context, ids []uint) error {
	return service.ErrInvalidLLMModelInput
}

func (s *contractLLMModelService) Search(ctx context.Context, query service.LLMModelSearchQuery) (*service.LLMModelPage, error) {
	return nil, service.ErrInvalidLLMModelInput
}

func (s *contractLLMModelService) Default(ctx context.Context) (*service.LLMModel, error) {
	return nil, service.ErrInvalidLLMModelInput
}

func (s *contractLLMModelService) SetDefault(ctx context.Context, id uint) error {
	return service.ErrInvalidLLMModelInput
}

func (s *contractLLMModelService) TestConnection(ctx context.Context, input service.LLMConnectionInput) error {
	return service.ErrInvalidLLMModelInput
}

type contractReviewLogService struct{}

func (s *contractReviewLogService) GetPush(ctx context.Context, id uint) (*service.PushReviewLog, error) {
	return nil, service.ErrInvalidReviewLogInput
}

func (s *contractReviewLogService) SearchPush(ctx context.Context, query service.ReviewLogSearchQuery) (*service.PushReviewLogPage, error) {
	return &service.PushReviewLogPage{
		Items: []service.PushReviewLog{},
		Total: 0,
		Page:  1,
		Size:  20,
	}, nil
}

func (s *contractReviewLogService) DeletePush(ctx context.Context, id uint) error {
	return service.ErrInvalidReviewLogInput
}

func (s *contractReviewLogService) PushAuthors(ctx context.Context, query service.ReviewLogOptionQuery) ([]service.AuthorOption, error) {
	return []service.AuthorOption{}, nil
}

func (s *contractReviewLogService) PushProjectNames(ctx context.Context, query service.ReviewLogOptionQuery) ([]string, error) {
	return []string{}, nil
}

func (s *contractReviewLogService) GeneratePushShareToken(ctx context.Context, id uint) (*service.ReviewLogShareToken, error) {
	return nil, service.ErrReviewLogNotFound
}

func (s *contractReviewLogService) GetMergeRequest(ctx context.Context, id uint) (*service.MergeRequestReviewLog, error) {
	return nil, service.ErrInvalidReviewLogInput
}

func (s *contractReviewLogService) SearchMergeRequest(ctx context.Context, query service.ReviewLogSearchQuery) (*service.MergeRequestReviewLogPage, error) {
	return &service.MergeRequestReviewLogPage{
		Items: []service.MergeRequestReviewLog{},
		Total: 0,
		Page:  1,
		Size:  20,
	}, nil
}

func (s *contractReviewLogService) DeleteMergeRequest(ctx context.Context, id uint) error {
	return service.ErrInvalidReviewLogInput
}

func (s *contractReviewLogService) MergeRequestAuthors(ctx context.Context, query service.ReviewLogOptionQuery) ([]service.AuthorOption, error) {
	return []service.AuthorOption{}, nil
}

func (s *contractReviewLogService) MergeRequestProjectNames(ctx context.Context, query service.ReviewLogOptionQuery) ([]string, error) {
	return []string{}, nil
}

func (s *contractReviewLogService) GenerateMergeRequestShareToken(ctx context.Context, id uint) (*service.ReviewLogShareToken, error) {
	return nil, service.ErrReviewLogNotFound
}

func (s *contractReviewLogService) GetShareToken(ctx context.Context, eventType string, eventID uint) (*service.ReviewLogShareToken, error) {
	return nil, service.ErrInvalidReviewLogInput
}
