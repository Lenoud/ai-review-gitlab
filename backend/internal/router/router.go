package router

import (
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/handler"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

type routeDef struct {
	method     string
	path       string
	permission string
}

type Dependencies struct {
	AuthHandler          *handler.AuthHandler
	ProjectHandler       *handler.ProjectHandler
	ProjectGitLabHandler *handler.ProjectGitLabHandler
	LLMModelHandler      *handler.LLMModelHandler
	ReviewLogHandler     *handler.ReviewLogHandler
	AIReviewTraceHandler *handler.AIReviewTraceHandler
	OpenReportHandler    *handler.OpenReportHandler
	WebhookHandler       *handler.WebhookHandler
	RBACHandler          *handler.RBACHandler
	AuthMiddleware       gin.HandlerFunc
}

func New(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestLogger())

	registerPublicRoutes(r, deps)
	registerAdminRoutes(r, deps)

	return r
}

func registerPublicRoutes(r *gin.Engine, deps Dependencies) {
	open := r.Group("/api/v1/open")
	{
		open.POST("/auth/login", deps.AuthHandler.Login)
		open.POST("/auth/refresh", deps.AuthHandler.Refresh)
		open.POST("/auth/logout", deps.AuthHandler.Logout)
		open.GET("/system/info", handler.SystemInfo)
		open.GET("/code-review-report", deps.OpenReportHandler.CodeReviewReport)
		open.GET("/analysis-report", deps.OpenReportHandler.AnalysisReport)
	}

	r.POST("/review/webhook", deps.WebhookHandler.GitLab)
}

func registerAdminRoutes(r *gin.Engine, deps Dependencies) {
	admin := r.Group("/api/v1/admin")
	admin.Use(deps.AuthMiddleware)
	admin.GET("/auth/me", deps.AuthHandler.Me)
	registerProjectRoutes(admin, deps.ProjectHandler)
	registerProjectGitLabRoutes(admin, deps.ProjectGitLabHandler)
	registerLLMModelRoutes(admin, deps.LLMModelHandler)
	registerReviewLogRoutes(admin, deps.ReviewLogHandler)
	registerAIReviewTraceRoutes(admin, deps.AIReviewTraceHandler)
	registerRBACRoutes(admin, deps.RBACHandler)
	registerRoutes(admin, []routeDef{
		{http.MethodPost, "/im-robot/create", "im-robot:write"},
		{http.MethodPost, "/im-robot/update", "im-robot:write"},
		{http.MethodGet, "/im-robot/get", "im-robot:read"},
		{http.MethodPost, "/im-robot/delete", "im-robot:write"},
		{http.MethodGet, "/im-robot/search", "im-robot:read"},
		{http.MethodGet, "/im-robot/list-enabled", "im-robot:read"},
		{http.MethodPost, "/im-robot/test-webhook", "im-robot:write"},

		{http.MethodPost, "/member-im-mapping/create", "member-im-mapping:write"},
		{http.MethodPost, "/member-im-mapping/update", "member-im-mapping:write"},
		{http.MethodGet, "/member-im-mapping/get", "member-im-mapping:read"},
		{http.MethodPost, "/member-im-mapping/delete", "member-im-mapping:write"},
		{http.MethodGet, "/member-im-mapping/search", "member-im-mapping:read"},

		{http.MethodPost, "/project-template/create", "project-template:write"},
		{http.MethodPost, "/project-template/update", "project-template:write"},
		{http.MethodGet, "/project-template/get", "project-template:read"},
		{http.MethodGet, "/project-template/list", "project-template:read"},
		{http.MethodPost, "/project-template/delete", "project-template:write"},
		{http.MethodGet, "/project-template/review-rule/list-by-template-id", "project-template:read"},
		{http.MethodPost, "/project-template/review-rule/create", "project-template:write"},
		{http.MethodPost, "/project-template/review-rule/update", "project-template:write"},
		{http.MethodGet, "/project-template/review-rule/get", "project-template:read"},
		{http.MethodPost, "/project-template/review-rule/delete", "project-template:write"},

		{http.MethodPost, "/project-analysis-plan/create", "project-analysis-plan:write"},
		{http.MethodPost, "/project-analysis-plan/update", "project-analysis-plan:write"},
		{http.MethodGet, "/project-analysis-plan/get", "project-analysis-plan:read"},
		{http.MethodPost, "/project-analysis-plan/delete", "project-analysis-plan:write"},
		{http.MethodGet, "/project-analysis-plan/search", "project-analysis-plan:read"},
		{http.MethodPost, "/project-analysis-plan-execution-log/test-run", "project-analysis-plan:write"},
		{http.MethodGet, "/project-analysis-plan-execution-log/get", "project-analysis-plan:read"},
		{http.MethodGet, "/project-analysis-plan-execution-log/search", "project-analysis-plan:read"},
		{http.MethodGet, "/project-analysis-plan-execution-log/html-report/:logId", "project-analysis-plan:read"},
		{http.MethodPost, "/project-analysis-plan-execution-log/generate-share-token/:logId", "project-analysis-plan:write"},

		{http.MethodPost, "/user/create", "rbac:write"},
		{http.MethodPost, "/user/update", "rbac:write"},
		{http.MethodGet, "/user/get", "rbac:read"},
		{http.MethodGet, "/user/search", "rbac:read"},
		{http.MethodGet, "/user/role-options", "rbac:read"},
		{http.MethodPost, "/role/create", "rbac:write"},
		{http.MethodPost, "/role/update", "rbac:write"},
		{http.MethodGet, "/role/get", "rbac:read"},
		{http.MethodPost, "/role/delete", "rbac:write"},

		{http.MethodGet, "/stats", "stats:read"},
		{http.MethodGet, "/member/commit-summary", "stats:read"},
		{http.MethodGet, "/sys-log/get", "sys-log:read"},
		{http.MethodGet, "/sys-log/search", "sys-log:read"},
		{http.MethodGet, "/system/info", "system:read"},
		{http.MethodGet, "/system/config", "system:read"},
		{http.MethodPost, "/system/config/base-url", "system:write"},
	})
}

func registerRBACRoutes(group *gin.RouterGroup, rbacHandler *handler.RBACHandler) {
	group.GET("/role/list", middleware.RequirePermission("rbac:read"), rbacHandler.ListRoles)
	group.GET("/role/menu-permissions", middleware.RequirePermission("rbac:read"), rbacHandler.MenuPermissions)
}

func registerReviewLogRoutes(group *gin.RouterGroup, reviewLogHandler *handler.ReviewLogHandler) {
	group.GET("/push-review-log/get", middleware.RequirePermission("review-log:read"), reviewLogHandler.GetPush)
	group.GET("/push-review-log/search", middleware.RequirePermission("review-log:read"), reviewLogHandler.SearchPush)
	group.POST("/push-review-log/delete", middleware.RequirePermission("review-log:write"), reviewLogHandler.DeletePush)
	group.GET("/push-review-log/authors", middleware.RequirePermission("review-log:read"), reviewLogHandler.PushAuthors)
	group.GET("/push-review-log/project-names", middleware.RequirePermission("review-log:read"), reviewLogHandler.PushProjectNames)
	group.POST("/push-review-log/generate-share-token/:logId", middleware.RequirePermission("review-log:write"), reviewLogHandler.GeneratePushShareToken)
	group.GET("/merge-request-review-log/get", middleware.RequirePermission("review-log:read"), reviewLogHandler.GetMergeRequest)
	group.GET("/merge-request-review-log/search", middleware.RequirePermission("review-log:read"), reviewLogHandler.SearchMergeRequest)
	group.POST("/merge-request-review-log/delete", middleware.RequirePermission("review-log:write"), reviewLogHandler.DeleteMergeRequest)
	group.GET("/merge-request-review-log/authors", middleware.RequirePermission("review-log:read"), reviewLogHandler.MergeRequestAuthors)
	group.GET("/merge-request-review-log/project-names", middleware.RequirePermission("review-log:read"), reviewLogHandler.MergeRequestProjectNames)
	group.POST("/merge-request-review-log/generate-share-token/:logId", middleware.RequirePermission("review-log:write"), reviewLogHandler.GenerateMergeRequestShareToken)
	group.GET("/review-log/get-share-token", middleware.RequirePermission("review-log:read"), reviewLogHandler.GetShareToken)
}

func registerAIReviewTraceRoutes(group *gin.RouterGroup, traceHandler *handler.AIReviewTraceHandler) {
	group.POST("/ai-review-trace/create", middleware.RequirePermission("ai-review-trace:write"), traceHandler.Create)
	group.GET("/ai-review-trace/get", middleware.RequirePermission("ai-review-trace:read"), traceHandler.Get)
}

func registerProjectRoutes(group *gin.RouterGroup, projectHandler *handler.ProjectHandler) {
	group.POST("/project/create", middleware.RequirePermission("project:write"), projectHandler.Create)
	group.POST("/project/batch-create", middleware.RequirePermission("project:write"), projectHandler.BatchCreate)
	group.POST("/project/update", middleware.RequirePermission("project:write"), projectHandler.Update)
	group.GET("/project/get", middleware.RequirePermission("project:read"), projectHandler.Get)
	group.POST("/project/delete", middleware.RequirePermission("project:write"), projectHandler.Delete)
	group.GET("/project/search", middleware.RequirePermission("project:read"), projectHandler.Search)
	group.POST("/project/web-urls/exists", middleware.RequirePermission("project:read"), projectHandler.WebURLExists)
	group.GET("/project/review-prompt/get", middleware.RequirePermission("project:read"), projectHandler.GetReviewPrompt)
	group.GET("/project/review-prompt/default", middleware.RequirePermission("project:read"), projectHandler.GetDefaultReviewPrompt)
	group.POST("/project/review-prompt/update", middleware.RequirePermission("project:write"), projectHandler.UpdateReviewPrompt)
	group.POST("/project/review-prompt/delete", middleware.RequirePermission("project:write"), projectHandler.DeleteReviewPrompt)
	group.POST("/project/review-prompt/test", middleware.RequirePermission("project:read"), projectHandler.TestReviewPrompt)
}

func registerProjectGitLabRoutes(group *gin.RouterGroup, gitLabHandler *handler.ProjectGitLabHandler) {
	group.POST("/project/gitlab/remote-search", middleware.RequirePermission("project:read"), gitLabHandler.RemoteSearch)
	group.POST("/project/gitlab/group-search", middleware.RequirePermission("project:read"), gitLabHandler.GroupSearch)
}

func registerLLMModelRoutes(group *gin.RouterGroup, llmModelHandler *handler.LLMModelHandler) {
	group.POST("/llm-model/create", middleware.RequirePermission("llm-model:write"), llmModelHandler.Create)
	group.POST("/llm-model/update", middleware.RequirePermission("llm-model:write"), llmModelHandler.Update)
	group.GET("/llm-model/get", middleware.RequirePermission("llm-model:read"), llmModelHandler.Get)
	group.POST("/llm-model/delete", middleware.RequirePermission("llm-model:write"), llmModelHandler.Delete)
	group.GET("/llm-model/search", middleware.RequirePermission("llm-model:read"), llmModelHandler.Search)
	group.GET("/llm-model/default", middleware.RequirePermission("llm-model:read"), llmModelHandler.Default)
	group.POST("/llm-model/set-default", middleware.RequirePermission("llm-model:write"), llmModelHandler.SetDefault)
	group.POST("/llm-test/connection", middleware.RequirePermission("llm-model:write"), llmModelHandler.TestConnection)
}

func registerRoutes(group *gin.RouterGroup, routes []routeDef) {
	for _, route := range routes {
		handlers := []gin.HandlerFunc{handler.NotImplemented}
		if route.permission != "" {
			handlers = append([]gin.HandlerFunc{middleware.RequirePermission(route.permission)}, handlers...)
		}
		switch route.method {
		case http.MethodGet:
			group.GET(route.path, handlers...)
		case http.MethodPost:
			group.POST(route.path, handlers...)
		}
	}
}
