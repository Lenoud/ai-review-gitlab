package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProjectGitLabService interface {
	SearchProjects(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabProject, error)
	SearchGroups(ctx context.Context, input service.GitLabSearchInput) ([]service.GitLabGroup, error)
}

type ProjectGitLabHandler struct {
	gitlab ProjectGitLabService
}

func NewProjectGitLabHandler(gitlab ProjectGitLabService) *ProjectGitLabHandler {
	return &ProjectGitLabHandler{gitlab: gitlab}
}

func (h *ProjectGitLabHandler) RemoteSearch(c *gin.Context) {
	var req gitLabSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "GitLab 参数错误")
		return
	}
	projects, err := h.gitlab.SearchProjects(c.Request.Context(), req.toInput())
	if err != nil {
		writeGitLabError(c, err)
		return
	}
	response.Success(c, projects)
}

func (h *ProjectGitLabHandler) GroupSearch(c *gin.Context) {
	var req gitLabSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "GitLab 参数错误")
		return
	}
	groups, err := h.gitlab.SearchGroups(c.Request.Context(), req.toInput())
	if err != nil {
		writeGitLabError(c, err)
		return
	}
	response.Success(c, groups)
}

func writeGitLabError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidGitLabInput):
		response.BadRequest(c, "GitLab 参数错误")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "GitLab 请求失败")
	}
}

type gitLabSearchRequest struct {
	APIBaseURL  string `json:"apiBaseUrl"`
	AccessToken string `json:"accessToken"`
	Keyword     string `json:"keyword"`
	Page        int    `json:"page"`
	Size        int    `json:"size"`
}

func (r gitLabSearchRequest) toInput() service.GitLabSearchInput {
	return service.GitLabSearchInput{
		APIBaseURL:  r.APIBaseURL,
		AccessToken: r.AccessToken,
		Keyword:     r.Keyword,
		Page:        r.Page,
		Size:        r.Size,
	}
}
