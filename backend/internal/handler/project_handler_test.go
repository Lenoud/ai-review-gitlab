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

type fakeProjectService struct {
	create       func(context.Context, service.ProjectInput) (*service.Project, error)
	batchCreate  func(context.Context, []service.ProjectInput) ([]service.Project, error)
	update       func(context.Context, uint, service.ProjectInput) (*service.Project, error)
	get          func(context.Context, uint) (*service.Project, error)
	delete       func(context.Context, []uint) error
	search       func(context.Context, service.ProjectSearchQuery) (*service.ProjectPage, error)
	webURLExists func(context.Context, string, uint) (bool, error)
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
