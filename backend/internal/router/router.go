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
	AuthHandler    *handler.AuthHandler
	AuthMiddleware gin.HandlerFunc
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
		open.GET("/code-review-report", handler.NotImplemented)
		open.GET("/analysis-report", handler.NotImplemented)
	}

	r.POST("/review/webhook", handler.NotImplemented)
}

func registerAdminRoutes(r *gin.Engine, deps Dependencies) {
	admin := r.Group("/api/v1/admin")
	admin.Use(deps.AuthMiddleware)
	admin.GET("/auth/me", deps.AuthHandler.Me)
	registerRoutes(admin, []routeDef{
		{http.MethodPost, "/project/create"},
		{http.MethodPost, "/project/batch-create"},
		{http.MethodPost, "/project/update"},
		{http.MethodGet, "/project/get"},
		{http.MethodPost, "/project/delete"},
		{http.MethodGet, "/project/search"},
		{http.MethodPost, "/project/gitlab/remote-search"},
		{http.MethodPost, "/project/gitlab/group-search"},
		{http.MethodPost, "/project/web-urls/exists"},
		{http.MethodGet, "/project/review-prompt/get"},
		{http.MethodGet, "/project/review-prompt/default"},
		{http.MethodPost, "/project/review-prompt/update"},
		{http.MethodPost, "/project/review-prompt/delete"},
		{http.MethodPost, "/project/review-prompt/test"},

		{http.MethodGet, "/push-review-log/get"},
		{http.MethodGet, "/push-review-log/search"},
		{http.MethodPost, "/push-review-log/delete"},
		{http.MethodGet, "/push-review-log/authors"},
		{http.MethodGet, "/push-review-log/project-names"},
		{http.MethodPost, "/push-review-log/generate-share-token/:logId"},
		{http.MethodGet, "/merge-request-review-log/get"},
		{http.MethodGet, "/merge-request-review-log/search"},
		{http.MethodPost, "/merge-request-review-log/delete"},
		{http.MethodGet, "/merge-request-review-log/authors"},
		{http.MethodGet, "/merge-request-review-log/project-names"},
		{http.MethodPost, "/merge-request-review-log/generate-share-token/:logId"},

		{http.MethodPost, "/ai-review-trace/create"},
		{http.MethodGet, "/ai-review-trace/get"},

		{http.MethodPost, "/llm-model/create"},
		{http.MethodPost, "/llm-model/update"},
		{http.MethodGet, "/llm-model/get"},
		{http.MethodPost, "/llm-model/delete"},
		{http.MethodGet, "/llm-model/search"},
		{http.MethodGet, "/llm-model/default"},
		{http.MethodPost, "/llm-model/set-default"},
		{http.MethodPost, "/llm-test/connection"},

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
		{http.MethodGet, "/review-log/get-share-token"},
	})
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
