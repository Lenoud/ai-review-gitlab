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
	AuthHandler                      *handler.AuthHandler
	ProjectHandler                   *handler.ProjectHandler
	ProjectGitLabHandler             *handler.ProjectGitLabHandler
	ProjectTemplateHandler           *handler.ProjectTemplateHandler
	ProjectTemplateReviewRuleHandler *handler.ProjectTemplateReviewRuleHandler
	AnalysisPlanHandler              *handler.ProjectAnalysisPlanHandler
	LLMModelHandler                  *handler.LLMModelHandler
	ReviewLogHandler                 *handler.ReviewLogHandler
	AnalysisLogHandler               *handler.AnalysisExecutionLogHandler
	AIReviewTraceHandler             *handler.AIReviewTraceHandler
	IMRobotHandler                   *handler.IMRobotHandler
	MemberIMMappingHandler           *handler.MemberIMMappingHandler
	OpenReportHandler                *handler.OpenReportHandler
	WebhookHandler                   *handler.WebhookHandler
	RBACHandler                      *handler.RBACHandler
	AuthMiddleware                   gin.HandlerFunc
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
	registerProjectTemplateRoutes(admin, deps.ProjectTemplateHandler)
	registerProjectTemplateReviewRuleRoutes(admin, deps.ProjectTemplateReviewRuleHandler)
	registerProjectAnalysisPlanRoutes(admin, deps.AnalysisPlanHandler)
	registerLLMModelRoutes(admin, deps.LLMModelHandler)
	registerReviewLogRoutes(admin, deps.ReviewLogHandler)
	registerAnalysisExecutionLogRoutes(admin, deps.AnalysisLogHandler)
	registerAIReviewTraceRoutes(admin, deps.AIReviewTraceHandler)
	registerIMRobotRoutes(admin, deps.IMRobotHandler)
	registerMemberIMMappingRoutes(admin, deps.MemberIMMappingHandler)
	registerRBACRoutes(admin, deps.RBACHandler)
	registerRoutes(admin, []routeDef{
		{http.MethodPost, "/im-robot/test-webhook", "im-robot:write"},

		{http.MethodPost, "/project-analysis-plan-execution-log/test-run", "project-analysis-plan:write"},

		{http.MethodGet, "/stats", "stats:read"},
		{http.MethodGet, "/member/commit-summary", "stats:read"},
		{http.MethodGet, "/sys-log/get", "sys-log:read"},
		{http.MethodGet, "/sys-log/search", "sys-log:read"},
		{http.MethodGet, "/system/info", "system:read"},
		{http.MethodGet, "/system/config", "system:read"},
		{http.MethodPost, "/system/config/base-url", "system:write"},
	})
}

func registerMemberIMMappingRoutes(group *gin.RouterGroup, mappingHandler *handler.MemberIMMappingHandler) {
	group.POST("/member-im-mapping/create", middleware.RequirePermission("member-im-mapping:write"), mappingHandler.Create)
	group.POST("/member-im-mapping/update", middleware.RequirePermission("member-im-mapping:write"), mappingHandler.Update)
	group.GET("/member-im-mapping/get", middleware.RequirePermission("member-im-mapping:read"), mappingHandler.Get)
	group.POST("/member-im-mapping/delete", middleware.RequirePermission("member-im-mapping:write"), mappingHandler.Delete)
	group.GET("/member-im-mapping/search", middleware.RequirePermission("member-im-mapping:read"), mappingHandler.Search)
}

func registerIMRobotRoutes(group *gin.RouterGroup, imRobotHandler *handler.IMRobotHandler) {
	group.POST("/im-robot/create", middleware.RequirePermission("im-robot:write"), imRobotHandler.Create)
	group.POST("/im-robot/update", middleware.RequirePermission("im-robot:write"), imRobotHandler.Update)
	group.GET("/im-robot/get", middleware.RequirePermission("im-robot:read"), imRobotHandler.Get)
	group.POST("/im-robot/delete", middleware.RequirePermission("im-robot:write"), imRobotHandler.Delete)
	group.GET("/im-robot/search", middleware.RequirePermission("im-robot:read"), imRobotHandler.Search)
	group.GET("/im-robot/list-enabled", middleware.RequirePermission("im-robot:read"), imRobotHandler.ListEnabled)
}

func registerProjectTemplateRoutes(group *gin.RouterGroup, templateHandler *handler.ProjectTemplateHandler) {
	group.POST("/project-template/create", middleware.RequirePermission("project-template:write"), templateHandler.Create)
	group.POST("/project-template/update", middleware.RequirePermission("project-template:write"), templateHandler.Update)
	group.GET("/project-template/get", middleware.RequirePermission("project-template:read"), templateHandler.Get)
	group.GET("/project-template/list", middleware.RequirePermission("project-template:read"), templateHandler.List)
	group.POST("/project-template/delete", middleware.RequirePermission("project-template:write"), templateHandler.Delete)
}

func registerProjectTemplateReviewRuleRoutes(group *gin.RouterGroup, ruleHandler *handler.ProjectTemplateReviewRuleHandler) {
	group.GET("/project-template/review-rule/list-by-template-id", middleware.RequirePermission("project-template:read"), ruleHandler.ListByTemplateID)
	group.POST("/project-template/review-rule/create", middleware.RequirePermission("project-template:write"), ruleHandler.Create)
	group.POST("/project-template/review-rule/update", middleware.RequirePermission("project-template:write"), ruleHandler.Update)
	group.GET("/project-template/review-rule/get", middleware.RequirePermission("project-template:read"), ruleHandler.Get)
	group.POST("/project-template/review-rule/delete", middleware.RequirePermission("project-template:write"), ruleHandler.Delete)
}

func registerProjectAnalysisPlanRoutes(group *gin.RouterGroup, analysisPlanHandler *handler.ProjectAnalysisPlanHandler) {
	group.POST("/project-analysis-plan/create", middleware.RequirePermission("project-analysis-plan:write"), analysisPlanHandler.Create)
	group.POST("/project-analysis-plan/update", middleware.RequirePermission("project-analysis-plan:write"), analysisPlanHandler.Update)
	group.GET("/project-analysis-plan/get", middleware.RequirePermission("project-analysis-plan:read"), analysisPlanHandler.Get)
	group.POST("/project-analysis-plan/delete", middleware.RequirePermission("project-analysis-plan:write"), analysisPlanHandler.Delete)
	group.GET("/project-analysis-plan/search", middleware.RequirePermission("project-analysis-plan:read"), analysisPlanHandler.Search)
}

func registerAnalysisExecutionLogRoutes(group *gin.RouterGroup, analysisLogHandler *handler.AnalysisExecutionLogHandler) {
	group.GET("/project-analysis-plan-execution-log/get", middleware.RequirePermission("project-analysis-plan:read"), analysisLogHandler.Get)
	group.GET("/project-analysis-plan-execution-log/search", middleware.RequirePermission("project-analysis-plan:read"), analysisLogHandler.Search)
	group.GET("/project-analysis-plan-execution-log/html-report/:logId", middleware.RequirePermission("project-analysis-plan:read"), analysisLogHandler.HTMLReport)
	group.POST("/project-analysis-plan-execution-log/generate-share-token/:logId", middleware.RequirePermission("project-analysis-plan:write"), analysisLogHandler.GenerateShareToken)
}

func registerRBACRoutes(group *gin.RouterGroup, rbacHandler *handler.RBACHandler) {
	group.POST("/user/create", middleware.RequirePermission("rbac:write"), rbacHandler.CreateUser)
	group.POST("/user/update", middleware.RequirePermission("rbac:write"), rbacHandler.UpdateUser)
	group.GET("/user/get", middleware.RequirePermission("rbac:read"), rbacHandler.GetUser)
	group.GET("/user/search", middleware.RequirePermission("rbac:read"), rbacHandler.SearchUsers)
	group.GET("/user/role-options", middleware.RequirePermission("rbac:read"), rbacHandler.RoleOptions)
	group.GET("/role/list", middleware.RequirePermission("rbac:read"), rbacHandler.ListRoles)
	group.POST("/role/create", middleware.RequirePermission("rbac:write"), rbacHandler.CreateRole)
	group.POST("/role/update", middleware.RequirePermission("rbac:write"), rbacHandler.UpdateRole)
	group.GET("/role/get", middleware.RequirePermission("rbac:read"), rbacHandler.GetRole)
	group.POST("/role/delete", middleware.RequirePermission("rbac:write"), rbacHandler.DeleteRole)
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
