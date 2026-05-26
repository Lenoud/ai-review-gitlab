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

func TestProjectHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectSvc := &fakeProjectService{
		create: func(ctx context.Context, input service.ProjectInput) (*service.Project, error) {
			require.Equal(t, "AI Review", input.Name)
			require.Equal(t, "https://gitlab.example.com/group/ai-review", input.WebURL)
			return &service.Project{ID: 1, Name: input.Name, WebURL: input.WebURL, Platform: input.Platform}, nil
		},
	}
	r := gin.New()
	r.POST("/project/create", NewProjectHandler(projectSvc).Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/project/create", strings.NewReader(`{"name":"AI Review","webUrl":"https://gitlab.example.com/group/ai-review","platform":"gitlab"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["id"])
	require.Equal(t, "AI Review", data["name"])
}

func TestProjectHandlerGetReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/project/get", NewProjectHandler(&fakeProjectService{
		get: func(ctx context.Context, id uint) (*service.Project, error) {
			return nil, service.ErrProjectNotFound
		},
	}).Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/project/get?id=99", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestProjectHandlerSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/project/search", NewProjectHandler(&fakeProjectService{
		search: func(ctx context.Context, query service.ProjectSearchQuery) (*service.ProjectPage, error) {
			require.Equal(t, "review", query.Keyword)
			require.Equal(t, 1, query.Page)
			require.Equal(t, 10, query.Size)
			return &service.ProjectPage{
				Items: []service.Project{{ID: 1, Name: "AI Review"}},
				Total: 1,
				Page:  1,
				Size:  10,
			}, nil
		},
	}).Search)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/project/search?keyword=review&page=1&size=10", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["total"])
}

func TestProjectHandlerWebURLExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/project/web-urls/exists", NewProjectHandler(&fakeProjectService{
		webURLExists: func(ctx context.Context, webURL string, excludeID uint) (bool, error) {
			require.Equal(t, "https://gitlab.example.com/group/ai-review", webURL)
			return true, nil
		},
	}).WebURLExists)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/project/web-urls/exists", strings.NewReader(`{"webUrl":"https://gitlab.example.com/group/ai-review"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, true, data["exists"])
}

func TestProjectHandlerReviewPromptRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectSvc := &fakeProjectService{
		getReviewPrompt: func(ctx context.Context, id uint) (*service.ReviewPrompt, error) {
			require.Equal(t, uint(1), id)
			return &service.ReviewPrompt{ProjectID: 1, PromptTemplate: "custom", Customized: true}, nil
		},
		getDefaultReviewPrompt: func(ctx context.Context) *service.ReviewPrompt {
			return &service.ReviewPrompt{PromptTemplate: "default"}
		},
		updateReviewPrompt: func(ctx context.Context, input service.ReviewPromptUpdateInput) (*service.ReviewPrompt, error) {
			require.Equal(t, uint(1), input.ProjectID)
			require.Equal(t, "custom", input.PromptTemplate)
			return &service.ReviewPrompt{ProjectID: 1, PromptTemplate: "custom", Customized: true}, nil
		},
		deleteReviewPrompt: func(ctx context.Context, id uint) error {
			require.Equal(t, uint(1), id)
			return nil
		},
		testReviewPrompt: func(ctx context.Context, input service.ReviewPromptTestInput) (*service.ReviewPromptTestResult, error) {
			require.Equal(t, uint(1), input.ProjectID)
			require.Equal(t, "项目 {{projectName}}", input.PromptTemplate)
			return &service.ReviewPromptTestResult{RenderedPrompt: "项目 AI Review", CharacterCount: len("项目 AI Review"), HasRequiredVariables: true}, nil
		},
	}
	handler := NewProjectHandler(projectSvc)
	r := gin.New()
	r.GET("/project/review-prompt/get", handler.GetReviewPrompt)
	r.GET("/project/review-prompt/default", handler.GetDefaultReviewPrompt)
	r.POST("/project/review-prompt/update", handler.UpdateReviewPrompt)
	r.POST("/project/review-prompt/delete", handler.DeleteReviewPrompt)
	r.POST("/project/review-prompt/test", handler.TestReviewPrompt)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/project/review-prompt/get?id=1", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/project/review-prompt/default", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/project/review-prompt/update", strings.NewReader(`{"id":1,"promptTemplate":"custom"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/project/review-prompt/delete", strings.NewReader(`{"id":1}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/project/review-prompt/test", strings.NewReader(`{"projectId":1,"promptTemplate":"项目 {{projectName}}"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, "项目 AI Review", data["renderedPrompt"])
}

type fakeProjectService struct {
	create                 func(context.Context, service.ProjectInput) (*service.Project, error)
	batchCreate            func(context.Context, []service.ProjectInput) ([]service.Project, error)
	update                 func(context.Context, uint, service.ProjectInput) (*service.Project, error)
	get                    func(context.Context, uint) (*service.Project, error)
	delete                 func(context.Context, []uint) error
	search                 func(context.Context, service.ProjectSearchQuery) (*service.ProjectPage, error)
	webURLExists           func(context.Context, string, uint) (bool, error)
	getReviewPrompt        func(context.Context, uint) (*service.ReviewPrompt, error)
	getDefaultReviewPrompt func(context.Context) *service.ReviewPrompt
	updateReviewPrompt     func(context.Context, service.ReviewPromptUpdateInput) (*service.ReviewPrompt, error)
	deleteReviewPrompt     func(context.Context, uint) error
	testReviewPrompt       func(context.Context, service.ReviewPromptTestInput) (*service.ReviewPromptTestResult, error)
}

func (s *fakeProjectService) Create(ctx context.Context, input service.ProjectInput) (*service.Project, error) {
	return s.create(ctx, input)
}
func (s *fakeProjectService) BatchCreate(ctx context.Context, inputs []service.ProjectInput) ([]service.Project, error) {
	return s.batchCreate(ctx, inputs)
}
func (s *fakeProjectService) Update(ctx context.Context, id uint, input service.ProjectInput) (*service.Project, error) {
	return s.update(ctx, id, input)
}
func (s *fakeProjectService) Get(ctx context.Context, id uint) (*service.Project, error) {
	return s.get(ctx, id)
}
func (s *fakeProjectService) Delete(ctx context.Context, ids []uint) error {
	return s.delete(ctx, ids)
}
func (s *fakeProjectService) Search(ctx context.Context, query service.ProjectSearchQuery) (*service.ProjectPage, error) {
	return s.search(ctx, query)
}
func (s *fakeProjectService) WebURLExists(ctx context.Context, webURL string, excludeID uint) (bool, error) {
	return s.webURLExists(ctx, webURL, excludeID)
}
func (s *fakeProjectService) GetReviewPrompt(ctx context.Context, id uint) (*service.ReviewPrompt, error) {
	return s.getReviewPrompt(ctx, id)
}
func (s *fakeProjectService) GetDefaultReviewPrompt(ctx context.Context) *service.ReviewPrompt {
	return s.getDefaultReviewPrompt(ctx)
}
func (s *fakeProjectService) UpdateReviewPrompt(ctx context.Context, input service.ReviewPromptUpdateInput) (*service.ReviewPrompt, error) {
	return s.updateReviewPrompt(ctx, input)
}
func (s *fakeProjectService) DeleteReviewPrompt(ctx context.Context, id uint) error {
	return s.deleteReviewPrompt(ctx, id)
}
func (s *fakeProjectService) TestReviewPrompt(ctx context.Context, input service.ReviewPromptTestInput) (*service.ReviewPromptTestResult, error) {
	return s.testReviewPrompt(ctx, input)
}
