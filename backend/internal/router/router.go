package router

import (
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/handler"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

type routeDef struct {
	method string
	path   string
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
		open.GET("/analysis-report", handler.NotImplemented)
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
	registerRoutes(admin, []routeDef{
		{http.MethodGet, "/project/review-prompt/get"},
		{http.MethodGet, "/project/review-prompt/default"},
		{http.MethodPost, "/project/review-prompt/update"},
		{http.MethodPost, "/project/review-prompt/delete"},
		{http.MethodPost, "/project/review-prompt/test"},

		{http.MethodPost, "/im-robot/create"},
		{http.MethodPost, "/im-robot/update"},
		{http.MethodGet, "/im-robot/get"},
		{http.MethodPost, "/im-robot/delete"},
		{http.MethodGet, "/im-robot/search"},
		{http.MethodGet, "/im-robot/list-enabled"},
		{http.MethodPost, "/im-robot/test-webhook"},

		{http.MethodPost, "/member-im-mapping/create"},
		{http.MethodPost, "/member-im-mapping/update"},
		{http.MethodGet, "/member-im-mapping/get"},
		{http.MethodPost, "/member-im-mapping/delete"},
		{http.MethodGet, "/member-im-mapping/search"},

		{http.MethodPost, "/project-template/create"},
		{http.MethodPost, "/project-template/update"},
		{http.MethodGet, "/project-template/get"},
		{http.MethodGet, "/project-template/list"},
		{http.MethodPost, "/project-template/delete"},
		{http.MethodGet, "/project-template/review-rule/list-by-template-id"},
		{http.MethodPost, "/project-template/review-rule/create"},
		{http.MethodPost, "/project-template/review-rule/update"},
		{http.MethodGet, "/project-template/review-rule/get"},
		{http.MethodPost, "/project-template/review-rule/delete"},

		{http.MethodPost, "/project-analysis-plan/create"},
		{http.MethodPost, "/project-analysis-plan/update"},
		{http.MethodGet, "/project-analysis-plan/get"},
		{http.MethodPost, "/project-analysis-plan/delete"},
		{http.MethodGet, "/project-analysis-plan/search"},
		{http.MethodPost, "/project-analysis-plan-execution-log/test-run"},
		{http.MethodGet, "/project-analysis-plan-execution-log/get"},
		{http.MethodGet, "/project-analysis-plan-execution-log/search"},
		{http.MethodGet, "/project-analysis-plan-execution-log/html-report/:logId"},
		{http.MethodPost, "/project-analysis-plan-execution-log/generate-share-token/:logId"},

		{http.MethodPost, "/user/create"},
		{http.MethodPost, "/user/update"},
		{http.MethodGet, "/user/get"},
		{http.MethodGet, "/user/search"},
		{http.MethodGet, "/user/role-options"},
		{http.MethodGet, "/role/list"},
		{http.MethodPost, "/role/create"},
		{http.MethodPost, "/role/update"},
		{http.MethodGet, "/role/get"},
		{http.MethodPost, "/role/delete"},
		{http.MethodGet, "/role/menu-permissions"},

		{http.MethodGet, "/stats"},
		{http.MethodGet, "/member/commit-summary"},
		{http.MethodGet, "/sys-log/get"},
		{http.MethodGet, "/sys-log/search"},
		{http.MethodGet, "/system/info"},
		{http.MethodGet, "/system/config"},
		{http.MethodPost, "/system/config/base-url"},
	})
}

func registerReviewLogRoutes(group *gin.RouterGroup, reviewLogHandler *handler.ReviewLogHandler) {
	group.GET("/push-review-log/get", reviewLogHandler.GetPush)
	group.GET("/push-review-log/search", reviewLogHandler.SearchPush)
	group.POST("/push-review-log/delete", reviewLogHandler.DeletePush)
	group.GET("/push-review-log/authors", reviewLogHandler.PushAuthors)
	group.GET("/push-review-log/project-names", reviewLogHandler.PushProjectNames)
	group.POST("/push-review-log/generate-share-token/:logId", reviewLogHandler.GeneratePushShareToken)
	group.GET("/merge-request-review-log/get", reviewLogHandler.GetMergeRequest)
	group.GET("/merge-request-review-log/search", reviewLogHandler.SearchMergeRequest)
	group.POST("/merge-request-review-log/delete", reviewLogHandler.DeleteMergeRequest)
	group.GET("/merge-request-review-log/authors", reviewLogHandler.MergeRequestAuthors)
	group.GET("/merge-request-review-log/project-names", reviewLogHandler.MergeRequestProjectNames)
	group.POST("/merge-request-review-log/generate-share-token/:logId", reviewLogHandler.GenerateMergeRequestShareToken)
	group.GET("/review-log/get-share-token", reviewLogHandler.GetShareToken)
}

func registerAIReviewTraceRoutes(group *gin.RouterGroup, traceHandler *handler.AIReviewTraceHandler) {
	group.POST("/ai-review-trace/create", traceHandler.Create)
	group.GET("/ai-review-trace/get", traceHandler.Get)
}

func registerProjectRoutes(group *gin.RouterGroup, projectHandler *handler.ProjectHandler) {
	group.POST("/project/create", projectHandler.Create)
	group.POST("/project/batch-create", projectHandler.BatchCreate)
	group.POST("/project/update", projectHandler.Update)
	group.GET("/project/get", projectHandler.Get)
	group.POST("/project/delete", projectHandler.Delete)
	group.GET("/project/search", projectHandler.Search)
	group.POST("/project/web-urls/exists", projectHandler.WebURLExists)
}

func registerProjectGitLabRoutes(group *gin.RouterGroup, gitLabHandler *handler.ProjectGitLabHandler) {
	group.POST("/project/gitlab/remote-search", gitLabHandler.RemoteSearch)
	group.POST("/project/gitlab/group-search", gitLabHandler.GroupSearch)
}

func registerLLMModelRoutes(group *gin.RouterGroup, llmModelHandler *handler.LLMModelHandler) {
	group.POST("/llm-model/create", llmModelHandler.Create)
	group.POST("/llm-model/update", llmModelHandler.Update)
	group.GET("/llm-model/get", llmModelHandler.Get)
	group.POST("/llm-model/delete", llmModelHandler.Delete)
	group.GET("/llm-model/search", llmModelHandler.Search)
	group.GET("/llm-model/default", llmModelHandler.Default)
	group.POST("/llm-model/set-default", llmModelHandler.SetDefault)
	group.POST("/llm-test/connection", llmModelHandler.TestConnection)
}

func registerRoutes(group *gin.RouterGroup, routes []routeDef) {
	for _, route := range routes {
		switch route.method {
		case http.MethodGet:
			group.GET(route.path, handler.NotImplemented)
		case http.MethodPost:
			group.POST(route.path, handler.NotImplemented)
		}
	}
}
