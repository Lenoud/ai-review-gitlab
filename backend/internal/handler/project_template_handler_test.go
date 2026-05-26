package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProjectTemplateHandlerCreateGetListAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewProjectTemplateHandler(&fakeProjectTemplateService{
		create: func(ctx context.Context, input service.ProjectTemplateInput) (*service.ProjectTemplate, error) {
			require.Equal(t, "Go service", input.Name)
			require.Equal(t, []string{".go"}, input.Extensions)
			return &service.ProjectTemplate{ID: 11, Name: "Go service", Extensions: input.Extensions}, nil
		},
		get: func(ctx context.Context, id uint) (*service.ProjectTemplate, error) {
			require.Equal(t, uint(11), id)
			return &service.ProjectTemplate{ID: 11, Name: "Go service"}, nil
		},
		list: func(ctx context.Context, query service.ProjectTemplateListQuery) ([]service.ProjectTemplate, error) {
			require.Equal(t, "Go", query.Keyword)
			return []service.ProjectTemplate{{ID: 11, Name: "Go service"}}, nil
		},
		delete: func(ctx context.Context, ids []uint) error {
			require.Equal(t, []uint{11}, ids)
			return nil
		},
	})
	r.POST("/create", handler.Create)
	r.GET("/get", handler.Get)
	r.GET("/list", handler.List)
	r.POST("/delete", handler.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(`{"name":"Go service","extensions":[".go"]}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/get?id=11", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/list?keyword=Go&page=2&size=10", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].([]any)
	require.Len(t, data, 1)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/delete", bytes.NewBufferString(`{"ids":[11]}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestProjectTemplateHandlerMapsErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewProjectTemplateHandler(&fakeProjectTemplateService{
		get: func(ctx context.Context, id uint) (*service.ProjectTemplate, error) {
			return nil, service.ErrProjectTemplateNotFound
		},
	})
	r.GET("/get", handler.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/get?id=404", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestProjectTemplateHandlerMapsTemplateInUse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewProjectTemplateHandler(&fakeProjectTemplateService{
		delete: func(ctx context.Context, ids []uint) error {
			return service.ErrProjectTemplateInUse
		},
	})
	r.POST("/delete", handler.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewBufferString(`{"ids":[7]}`))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

type fakeProjectTemplateService struct {
	create func(context.Context, service.ProjectTemplateInput) (*service.ProjectTemplate, error)
	update func(context.Context, uint, service.ProjectTemplateInput) (*service.ProjectTemplate, error)
	get    func(context.Context, uint) (*service.ProjectTemplate, error)
	delete func(context.Context, []uint) error
	list   func(context.Context, service.ProjectTemplateListQuery) ([]service.ProjectTemplate, error)
}

func (s *fakeProjectTemplateService) Create(ctx context.Context, input service.ProjectTemplateInput) (*service.ProjectTemplate, error) {
	return s.create(ctx, input)
}

func (s *fakeProjectTemplateService) Update(ctx context.Context, id uint, input service.ProjectTemplateInput) (*service.ProjectTemplate, error) {
	return s.update(ctx, id, input)
}

func (s *fakeProjectTemplateService) Get(ctx context.Context, id uint) (*service.ProjectTemplate, error) {
	return s.get(ctx, id)
}

func (s *fakeProjectTemplateService) Delete(ctx context.Context, ids []uint) error {
	return s.delete(ctx, ids)
}

func (s *fakeProjectTemplateService) List(ctx context.Context, query service.ProjectTemplateListQuery) ([]service.ProjectTemplate, error) {
	return s.list(ctx, query)
}
