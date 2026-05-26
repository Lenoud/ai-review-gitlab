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
		case "/api/v1/admin/role/list",
			"/api/v1/admin/role/menu-permissions":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/role/create",
			"/api/v1/admin/role/update",
			"/api/v1/admin/role/get",
			"/api/v1/admin/role/delete":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/user/create",
			"/api/v1/admin/user/update",
			"/api/v1/admin/user/get":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/user/search",
			"/api/v1/admin/user/role-options":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/project/create",
			"/api/v1/admin/project/batch-create",
			"/api/v1/admin/project/update",
			"/api/v1/admin/project/get",
			"/api/v1/admin/project/delete",
			"/api/v1/admin/project/search",
			"/api/v1/admin/project/gitlab/remote-search",
			"/api/v1/admin/project/gitlab/group-search",
			"/api/v1/admin/project/web-urls/exists",
			"/api/v1/admin/project/review-prompt/get",
			"/api/v1/admin/project/review-prompt/update",
			"/api/v1/admin/project/review-prompt/delete",
			"/api/v1/admin/project/review-prompt/test":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/project/review-prompt/default":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/llm-model/create",
			"/api/v1/admin/llm-model/update",
			"/api/v1/admin/llm-model/get",
			"/api/v1/admin/llm-model/delete",
			"/api/v1/admin/llm-model/search",
			"/api/v1/admin/llm-model/default",
			"/api/v1/admin/llm-model/set-default",
			"/api/v1/admin/llm-test/connection":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/im-robot/create",
			"/api/v1/admin/im-robot/update",
			"/api/v1/admin/im-robot/get",
			"/api/v1/admin/im-robot/delete":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/im-robot/search",
			"/api/v1/admin/im-robot/list-enabled":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/member-im-mapping/create",
			"/api/v1/admin/member-im-mapping/update",
			"/api/v1/admin/member-im-mapping/get",
			"/api/v1/admin/member-im-mapping/delete":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/member-im-mapping/search":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/push-review-log/get",
			"/api/v1/admin/merge-request-review-log/get",
			"/api/v1/admin/review-log/get-share-token",
			"/api/v1/admin/ai-review-trace/create",
			"/api/v1/admin/ai-review-trace/get":
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
		case "/api/v1/admin/project-analysis-plan-execution-log/get":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/project-analysis-plan-execution-log/search":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/project-analysis-plan-execution-log/html-report/1":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/project-analysis-plan-execution-log/generate-share-token/1":
			expectedStatus = http.StatusNotFound
		case "/api/v1/admin/project-analysis-plan/create",
			"/api/v1/admin/project-analysis-plan/update",
			"/api/v1/admin/project-analysis-plan/get",
			"/api/v1/admin/project-analysis-plan/delete":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/project-analysis-plan/search":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/project-template/create",
			"/api/v1/admin/project-template/update",
			"/api/v1/admin/project-template/get",
			"/api/v1/admin/project-template/delete":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/project-template/list":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/project-template/review-rule/create",
			"/api/v1/admin/project-template/review-rule/update",
			"/api/v1/admin/project-template/review-rule/get",
			"/api/v1/admin/project-template/review-rule/list-by-template-id",
			"/api/v1/admin/project-template/review-rule/delete":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/system/info",
			"/api/v1/admin/system/config":
			expectedStatus = http.StatusOK
		case "/api/v1/admin/system/config/base-url":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/stats",
			"/api/v1/admin/member/commit-summary":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/sys-log/get":
			expectedStatus = http.StatusBadRequest
		case "/api/v1/admin/sys-log/search":
			expectedStatus = http.StatusOK
		}
		require.Equal(t, expectedStatus, w.Code, "%s %s with token", route.method, route.path)
	}
}

func TestAdminProtectedRoutesReturnForbiddenWithoutPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New(Dependencies{
		AuthHandler:                      NewAuthHandlerForTest(),
		ProjectHandler:                   handler.NewProjectHandler(&contractProjectService{}),
		ProjectGitLabHandler:             handler.NewProjectGitLabHandler(&contractProjectGitLabService{}),
		ProjectTemplateHandler:           handler.NewProjectTemplateHandler(&contractProjectTemplateService{}),
		ProjectTemplateReviewRuleHandler: handler.NewProjectTemplateReviewRuleHandler(&contractProjectTemplateReviewRuleService{}),
		AnalysisPlanHandler:              handler.NewProjectAnalysisPlanHandler(&contractProjectAnalysisPlanService{}),
		LLMModelHandler:                  handler.NewLLMModelHandler(&contractLLMModelService{}),
		ReviewLogHandler:                 handler.NewReviewLogHandler(&contractReviewLogService{}),
		AnalysisLogHandler:               handler.NewAnalysisExecutionLogHandler(&contractAnalysisExecutionLogService{}),
		AIReviewTraceHandler:             handler.NewAIReviewTraceHandler(&contractAIReviewTraceService{}),
		IMRobotHandler:                   handler.NewIMRobotHandler(&contractIMRobotService{}),
		MemberIMMappingHandler:           handler.NewMemberIMMappingHandler(&contractMemberIMMappingService{}),
		SystemHandler:                    handler.NewSystemHandler(&contractSystemService{}),
		StatsHandler:                     handler.NewStatsHandler(&contractStatsService{}),
		SysLogHandler:                    handler.NewSysLogHandler(&contractSysLogService{}),
		OpenReportHandler:                handler.NewOpenReportHandler(&contractOpenReportService{}),
		WebhookHandler:                   handler.NewWebhookHandler(&contractReviewTaskService{}),
		RBACHandler:                      handler.NewRBACHandler(&contractRBACService{}),
		AuthMiddleware: middleware.JWTAuth(&contractTokenValidator{
			subject: &service.AuthSubject{
				UserID:   2,
				Username: "reviewer",
				Nickname: "Reviewer",
			},
		}),
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/project/create", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer access-token")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func newContractRouter() *gin.Engine {
	return New(Dependencies{
		AuthHandler:                      NewAuthHandlerForTest(),
		ProjectHandler:                   handler.NewProjectHandler(&contractProjectService{}),
		ProjectGitLabHandler:             handler.NewProjectGitLabHandler(&contractProjectGitLabService{}),
		ProjectTemplateHandler:           handler.NewProjectTemplateHandler(&contractProjectTemplateService{}),
		ProjectTemplateReviewRuleHandler: handler.NewProjectTemplateReviewRuleHandler(&contractProjectTemplateReviewRuleService{}),
		AnalysisPlanHandler:              handler.NewProjectAnalysisPlanHandler(&contractProjectAnalysisPlanService{}),
		LLMModelHandler:                  handler.NewLLMModelHandler(&contractLLMModelService{}),
		ReviewLogHandler:                 handler.NewReviewLogHandler(&contractReviewLogService{}),
		AnalysisLogHandler:               handler.NewAnalysisExecutionLogHandler(&contractAnalysisExecutionLogService{}),
		AIReviewTraceHandler:             handler.NewAIReviewTraceHandler(&contractAIReviewTraceService{}),
		IMRobotHandler:                   handler.NewIMRobotHandler(&contractIMRobotService{}),
		MemberIMMappingHandler:           handler.NewMemberIMMappingHandler(&contractMemberIMMappingService{}),
		SystemHandler:                    handler.NewSystemHandler(&contractSystemService{}),
		StatsHandler:                     handler.NewStatsHandler(&contractStatsService{}),
		SysLogHandler:                    handler.NewSysLogHandler(&contractSysLogService{}),
		OpenReportHandler:                handler.NewOpenReportHandler(&contractOpenReportService{}),
		WebhookHandler:                   handler.NewWebhookHandler(&contractReviewTaskService{}),
		RBACHandler:                      handler.NewRBACHandler(&contractRBACService{}),
		AuthMiddleware: middleware.JWTAuth(&contractTokenValidator{
			subject: &service.AuthSubject{
				UserID:   1,
				Username: "admin",
				Nickname: "Administrator",
				Roles:    []string{"admin"},
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

type contractRBACService struct{}

func (s *contractRBACService) ListRoles(ctx context.Context) ([]service.Role, error) {
	return []service.Role{}, nil
}

func (s *contractRBACService) ListPermissionGroups(ctx context.Context) ([]service.PermissionGroup, error) {
	return []service.PermissionGroup{}, nil
}

func (s *contractRBACService) CreateRole(ctx context.Context, input service.RoleInput) (*service.RoleDetail, error) {
	return nil, service.ErrInvalidRBACInput
}

func (s *contractRBACService) UpdateRole(ctx context.Context, id uint, input service.RoleInput) (*service.RoleDetail, error) {
	return nil, service.ErrInvalidRBACInput
}

func (s *contractRBACService) GetRole(ctx context.Context, id uint) (*service.RoleDetail, error) {
	return nil, service.ErrInvalidRBACInput
}

func (s *contractRBACService) DeleteRoles(ctx context.Context, ids []uint) error {
	return service.ErrInvalidRBACInput
}

func (s *contractRBACService) CreateUser(ctx context.Context, input service.AdminUserInput) (*service.AdminUser, error) {
	return nil, service.ErrInvalidRBACInput
}

func (s *contractRBACService) UpdateUser(ctx context.Context, id uint, input service.AdminUserInput) (*service.AdminUser, error) {
	return nil, service.ErrInvalidRBACInput
}

func (s *contractRBACService) GetUser(ctx context.Context, id uint) (*service.AdminUser, error) {
	return nil, service.ErrInvalidRBACInput
}

func (s *contractRBACService) SearchUsers(ctx context.Context, query service.AdminUserSearchQuery) (*service.AdminUserPage, error) {
	return &service.AdminUserPage{Items: []service.AdminUser{}, Page: query.Page, Size: query.Size}, nil
}

func (s *contractRBACService) ListRoleOptions(ctx context.Context) ([]service.Role, error) {
	return []service.Role{}, nil
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

func (s *contractProjectService) GetReviewPrompt(ctx context.Context, id uint) (*service.ReviewPrompt, error) {
	return nil, service.ErrInvalidProjectInput
}

func (s *contractProjectService) GetDefaultReviewPrompt(ctx context.Context) *service.ReviewPrompt {
	return &service.ReviewPrompt{PromptTemplate: "default"}
}

func (s *contractProjectService) UpdateReviewPrompt(ctx context.Context, input service.ReviewPromptUpdateInput) (*service.ReviewPrompt, error) {
	return nil, service.ErrInvalidProjectInput
}

func (s *contractProjectService) DeleteReviewPrompt(ctx context.Context, id uint) error {
	return service.ErrInvalidProjectInput
}

func (s *contractProjectService) TestReviewPrompt(ctx context.Context, input service.ReviewPromptTestInput) (*service.ReviewPromptTestResult, error) {
	return nil, service.ErrInvalidProjectInput
}

type contractProjectGitLabService struct{}

func (s *contractProjectGitLabService) SearchProjects(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabProject, error) {
	return nil, service.ErrInvalidGitLabInput
}

func (s *contractProjectGitLabService) SearchGroups(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabGroup, error) {
	return nil, service.ErrInvalidGitLabInput
}

type contractProjectTemplateService struct{}

func (s *contractProjectTemplateService) Create(ctx context.Context, input service.ProjectTemplateInput) (*service.ProjectTemplate, error) {
	return nil, service.ErrInvalidProjectTemplateInput
}

func (s *contractProjectTemplateService) Update(ctx context.Context, id uint, input service.ProjectTemplateInput) (*service.ProjectTemplate, error) {
	return nil, service.ErrInvalidProjectTemplateInput
}

func (s *contractProjectTemplateService) Get(ctx context.Context, id uint) (*service.ProjectTemplate, error) {
	return nil, service.ErrInvalidProjectTemplateInput
}

func (s *contractProjectTemplateService) Delete(ctx context.Context, ids []uint) error {
	return service.ErrInvalidProjectTemplateInput
}

func (s *contractProjectTemplateService) List(ctx context.Context, query service.ProjectTemplateListQuery) ([]service.ProjectTemplate, error) {
	return []service.ProjectTemplate{}, nil
}

type contractProjectTemplateReviewRuleService struct{}

func (s *contractProjectTemplateReviewRuleService) Create(ctx context.Context, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error) {
	return nil, service.ErrInvalidProjectTemplateReviewRuleInput
}

func (s *contractProjectTemplateReviewRuleService) Update(ctx context.Context, id uint, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error) {
	return nil, service.ErrInvalidProjectTemplateReviewRuleInput
}

func (s *contractProjectTemplateReviewRuleService) Get(ctx context.Context, id uint) (*service.ProjectTemplateReviewRule, error) {
	return nil, service.ErrInvalidProjectTemplateReviewRuleInput
}

func (s *contractProjectTemplateReviewRuleService) Delete(ctx context.Context, id uint) error {
	return service.ErrInvalidProjectTemplateReviewRuleInput
}

func (s *contractProjectTemplateReviewRuleService) ListByTemplateID(ctx context.Context, templateID uint) ([]service.ProjectTemplateReviewRule, error) {
	return nil, service.ErrInvalidProjectTemplateReviewRuleInput
}

type contractProjectAnalysisPlanService struct{}

func (s *contractProjectAnalysisPlanService) Create(ctx context.Context, input service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error) {
	return nil, service.ErrInvalidProjectAnalysisPlanInput
}

func (s *contractProjectAnalysisPlanService) Update(ctx context.Context, id uint, input service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error) {
	return nil, service.ErrInvalidProjectAnalysisPlanInput
}

func (s *contractProjectAnalysisPlanService) Get(ctx context.Context, id uint) (*service.ProjectAnalysisPlan, error) {
	return nil, service.ErrInvalidProjectAnalysisPlanInput
}

func (s *contractProjectAnalysisPlanService) Delete(ctx context.Context, ids []uint) error {
	return service.ErrInvalidProjectAnalysisPlanInput
}

func (s *contractProjectAnalysisPlanService) Search(ctx context.Context, query service.ProjectAnalysisPlanSearchQuery) (*service.ProjectAnalysisPlanPage, error) {
	return &service.ProjectAnalysisPlanPage{
		Items: []service.ProjectAnalysisPlan{},
		Total: 0,
		Page:  1,
		Size:  20,
	}, nil
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

type contractIMRobotService struct{}

func (s *contractIMRobotService) Create(ctx context.Context, input service.IMRobotInput) (*service.IMRobot, error) {
	return nil, service.ErrInvalidIMRobotInput
}

func (s *contractIMRobotService) Update(ctx context.Context, id uint, input service.IMRobotInput) (*service.IMRobot, error) {
	return nil, service.ErrInvalidIMRobotInput
}

func (s *contractIMRobotService) Get(ctx context.Context, id uint) (*service.IMRobot, error) {
	return nil, service.ErrInvalidIMRobotInput
}

func (s *contractIMRobotService) Delete(ctx context.Context, ids []uint) error {
	return service.ErrInvalidIMRobotInput
}

func (s *contractIMRobotService) Search(ctx context.Context, query service.IMRobotSearchQuery) (*service.IMRobotPage, error) {
	return &service.IMRobotPage{Items: []service.IMRobot{}, Total: 0, Page: 1, Size: 20}, nil
}

func (s *contractIMRobotService) ListEnabled(ctx context.Context) ([]service.IMRobot, error) {
	return []service.IMRobot{}, nil
}

type contractMemberIMMappingService struct{}

func (s *contractMemberIMMappingService) Create(ctx context.Context, input service.MemberIMMappingInput) (*service.MemberIMMapping, error) {
	return nil, service.ErrInvalidMemberIMMappingInput
}

func (s *contractMemberIMMappingService) Update(ctx context.Context, id uint, input service.MemberIMMappingInput) (*service.MemberIMMapping, error) {
	return nil, service.ErrInvalidMemberIMMappingInput
}

func (s *contractMemberIMMappingService) Get(ctx context.Context, id uint) (*service.MemberIMMapping, error) {
	return nil, service.ErrInvalidMemberIMMappingInput
}

func (s *contractMemberIMMappingService) Delete(ctx context.Context, ids []uint) error {
	return service.ErrInvalidMemberIMMappingInput
}

func (s *contractMemberIMMappingService) Search(ctx context.Context, query service.MemberIMMappingSearchQuery) (*service.MemberIMMappingPage, error) {
	return &service.MemberIMMappingPage{Items: []service.MemberIMMapping{}, Total: 0, Page: 1, Size: 20}, nil
}

type contractSystemService struct{}

func (s *contractSystemService) GetConfig(ctx context.Context) (*service.SystemConfig, error) {
	return &service.SystemConfig{Version: "1.0.0", SiteName: "AI Code Review", SiteNotice: ""}, nil
}

func (s *contractSystemService) UpdateBaseURL(ctx context.Context, baseURL string) (*service.SystemConfig, error) {
	return nil, service.ErrInvalidSystemConfigInput
}

type contractStatsService struct{}

func (s *contractStatsService) GetStats(ctx context.Context, query service.StatsRange) (*service.StatsOverview, error) {
	return &service.StatsOverview{}, nil
}

func (s *contractStatsService) GetMemberCommitSummary(ctx context.Context, query service.MemberCommitSummaryQuery) (*service.MemberCommitStatsPage, error) {
	return &service.MemberCommitStatsPage{Items: []service.MemberCommitStats{}, Page: 1, Size: 20}, nil
}

type contractSysLogService struct{}

func (s *contractSysLogService) Get(ctx context.Context, id uint) (*service.SysLog, error) {
	return nil, service.ErrInvalidSysLogInput
}

func (s *contractSysLogService) Search(ctx context.Context, query service.SysLogSearchQuery) (*service.SysLogPage, error) {
	return &service.SysLogPage{Items: []service.SysLog{}, Page: 1, Size: 20}, nil
}

type contractReviewTaskService struct{}

func (s *contractReviewTaskService) EnqueueGitLabWebhook(ctx context.Context, input service.GitLabWebhookInput) (*service.ReviewTaskEnqueueResult, error) {
	return nil, service.ErrInvalidReviewTaskInput
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

type contractAnalysisExecutionLogService struct{}

func (s *contractAnalysisExecutionLogService) Get(ctx context.Context, id uint) (*service.ProjectAnalysisPlanExecutionLog, error) {
	if id == 1 {
		return &service.ProjectAnalysisPlanExecutionLog{
			ID:            1,
			ProjectID:     7,
			PlanID:        3,
			Status:        "succeeded",
			ResultContent: "analysis",
		}, nil
	}
	return nil, service.ErrInvalidReviewLogInput
}

func (s *contractAnalysisExecutionLogService) Search(ctx context.Context, query service.AnalysisExecutionLogSearchQuery) (*service.AnalysisExecutionLogPage, error) {
	return &service.AnalysisExecutionLogPage{
		Items: []service.ProjectAnalysisPlanExecutionLog{},
		Total: 0,
		Page:  1,
		Size:  20,
	}, nil
}

func (s *contractAnalysisExecutionLogService) GenerateShareToken(ctx context.Context, id uint) (*service.ReviewLogShareToken, error) {
	return nil, service.ErrReviewLogNotFound
}

type contractOpenReportService struct{}

func (s *contractOpenReportService) CodeReviewReport(ctx context.Context, input service.CodeReviewReportInput) (string, error) {
	return "", service.ErrInvalidReviewLogInput
}

func (s *contractOpenReportService) AnalysisReport(ctx context.Context, input service.AnalysisReportInput) (string, error) {
	return "", service.ErrInvalidReviewLogInput
}

type contractAIReviewTraceService struct{}

func (s *contractAIReviewTraceService) Create(ctx context.Context, input service.AIReviewTraceInput) (*service.AIReviewTrace, error) {
	return nil, service.ErrInvalidAIReviewTraceInput
}

func (s *contractAIReviewTraceService) Get(ctx context.Context, eventType string, eventID uint) (*service.AIReviewTrace, error) {
	return nil, service.ErrInvalidAIReviewTraceInput
}
