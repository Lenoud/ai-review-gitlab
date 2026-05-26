package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProjectService interface {
	Create(ctx context.Context, input service.ProjectInput) (*service.Project, error)
	BatchCreate(ctx context.Context, inputs []service.ProjectInput) ([]service.Project, error)
	Update(ctx context.Context, id uint, input service.ProjectInput) (*service.Project, error)
	Get(ctx context.Context, id uint) (*service.Project, error)
	Delete(ctx context.Context, ids []uint) error
	Search(ctx context.Context, query service.ProjectSearchQuery) (*service.ProjectPage, error)
	WebURLExists(ctx context.Context, webURL string, excludeID uint) (bool, error)
	GetReviewPrompt(ctx context.Context, id uint) (*service.ReviewPrompt, error)
	GetDefaultReviewPrompt(ctx context.Context) *service.ReviewPrompt
	UpdateReviewPrompt(ctx context.Context, input service.ReviewPromptUpdateInput) (*service.ReviewPrompt, error)
	DeleteReviewPrompt(ctx context.Context, id uint) error
	TestReviewPrompt(ctx context.Context, input service.ReviewPromptTestInput) (*service.ReviewPromptTestResult, error)
}

type ProjectHandler struct {
	projects ProjectService
}

func NewProjectHandler(projects ProjectService) *ProjectHandler {
	return &ProjectHandler{projects: projects}
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req projectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "项目参数错误")
		return
	}
	project, err := h.projects.Create(c.Request.Context(), req.toInput())
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, project)
}

func (h *ProjectHandler) BatchCreate(c *gin.Context) {
	var req batchProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "项目参数错误")
		return
	}
	inputs := make([]service.ProjectInput, 0, len(req.Projects))
	for _, project := range req.Projects {
		inputs = append(inputs, project.toInput())
	}
	projects, err := h.projects.BatchCreate(c.Request.Context(), inputs)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, projects)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	var req projectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "项目参数错误")
		return
	}
	project, err := h.projects.Update(c.Request.Context(), req.ID, req.toInput())
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, project)
}

func (h *ProjectHandler) Get(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "项目ID不能为空")
		return
	}
	project, err := h.projects.Get(c.Request.Context(), id)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, project)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择要删除的项目")
		return
	}
	if err := h.projects.Delete(c.Request.Context(), req.IDs); err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ProjectHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := h.projects.Search(c.Request.Context(), service.ProjectSearchQuery{
		Keyword:  c.Query("keyword"),
		Platform: c.Query("platform"),
		Page:     page,
		Size:     size,
	})
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProjectHandler) WebURLExists(c *gin.Context) {
	var req webURLExistsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "项目 URL 不能为空")
		return
	}
	exists, err := h.projects.WebURLExists(c.Request.Context(), req.WebURL, req.ExcludeID)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, gin.H{"exists": exists})
}

func (h *ProjectHandler) GetReviewPrompt(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "项目ID不能为空")
		return
	}
	prompt, err := h.projects.GetReviewPrompt(c.Request.Context(), id)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, prompt)
}

func (h *ProjectHandler) GetDefaultReviewPrompt(c *gin.Context) {
	response.Success(c, h.projects.GetDefaultReviewPrompt(c.Request.Context()))
}

func (h *ProjectHandler) UpdateReviewPrompt(c *gin.Context) {
	var req reviewPromptUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Review提示词参数错误")
		return
	}
	prompt, err := h.projects.UpdateReviewPrompt(c.Request.Context(), service.ReviewPromptUpdateInput{
		ProjectID:      req.ID,
		PromptTemplate: req.PromptTemplate,
	})
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, prompt)
}

func (h *ProjectHandler) DeleteReviewPrompt(c *gin.Context) {
	var req idRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ID == 0 {
		response.BadRequest(c, "项目ID不能为空")
		return
	}
	if err := h.projects.DeleteReviewPrompt(c.Request.Context(), req.ID); err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *ProjectHandler) TestReviewPrompt(c *gin.Context) {
	var req reviewPromptTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Review提示词参数错误")
		return
	}
	result, err := h.projects.TestReviewPrompt(c.Request.Context(), service.ReviewPromptTestInput{
		ProjectID:      req.ProjectID,
		PromptTemplate: req.PromptTemplate,
		Diffs:          req.Diffs,
		Commits:        req.Commits,
	})
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.Success(c, result)
}

func writeProjectError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidProjectInput):
		response.BadRequest(c, "项目参数错误")
	case errors.Is(err, service.ErrProjectWebURLExists):
		response.BadRequest(c, "项目 URL 已存在")
	case errors.Is(err, service.ErrProjectNotFound):
		response.Error(c, http.StatusNotFound, 40400, "项目不存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "项目操作失败")
	}
}

type projectRequest struct {
	ID                       uint     `json:"id"`
	Name                     string   `json:"name"`
	Description              string   `json:"description"`
	WebURL                   string   `json:"webUrl"`
	Platform                 string   `json:"platform"`
	AccessToken              string   `json:"accessToken"`
	IMEnabled                bool     `json:"imEnabled"`
	IMRobotID                uint     `json:"imRobotId"`
	IMAtMemberEnabled        bool     `json:"imAtMemberEnabled"`
	IMAtMemberScoreThreshold int      `json:"imAtMemberScoreThreshold"`
	AIReviewEnabled          *bool    `json:"aiReviewEnabled"`
	TemplateID               uint     `json:"templateId"`
	Extensions               []string `json:"extensions"`
	ReviewEventTypes         []string `json:"reviewEventTypes"`
	ReviewPromptTemplate     string   `json:"reviewPromptTemplate"`
	HTMLReportEnabled        bool     `json:"htmlReportEnabled"`
	DeepReviewEnabled        bool     `json:"deepReviewEnabled"`
}

func (r projectRequest) toInput() service.ProjectInput {
	return service.ProjectInput{
		Name:                     r.Name,
		Description:              r.Description,
		WebURL:                   r.WebURL,
		Platform:                 r.Platform,
		AccessToken:              r.AccessToken,
		IMEnabled:                r.IMEnabled,
		IMRobotID:                r.IMRobotID,
		IMAtMemberEnabled:        r.IMAtMemberEnabled,
		IMAtMemberScoreThreshold: r.IMAtMemberScoreThreshold,
		AIReviewEnabled:          r.AIReviewEnabled,
		TemplateID:               r.TemplateID,
		Extensions:               r.Extensions,
		ReviewEventTypes:         r.ReviewEventTypes,
		ReviewPromptTemplate:     r.ReviewPromptTemplate,
		HTMLReportEnabled:        r.HTMLReportEnabled,
		DeepReviewEnabled:        r.DeepReviewEnabled,
	}
}

type batchProjectCreateRequest struct {
	Projects []projectRequest `json:"projects"`
}

type deleteRequest struct {
	IDs []uint `json:"ids"`
}

type idRequest struct {
	ID uint `json:"id"`
}

type webURLExistsRequest struct {
	WebURL    string `json:"webUrl"`
	ExcludeID uint   `json:"excludeId"`
}

type reviewPromptUpdateRequest struct {
	ID             uint   `json:"id"`
	PromptTemplate string `json:"promptTemplate"`
}

type reviewPromptTestRequest struct {
	ProjectID      uint   `json:"projectId"`
	PromptTemplate string `json:"promptTemplate"`
	Diffs          string `json:"diffs"`
	Commits        string `json:"commits"`
}

func parseUintQuery(c *gin.Context, key string) (uint, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, false
	}
	return uint(parsed), true
}
