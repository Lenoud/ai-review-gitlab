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

func TestProjectAnalysisPlanHandlerCreateGetSearchAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewProjectAnalysisPlanHandler(&fakeProjectAnalysisPlanService{
		create: func(ctx context.Context, input service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error) {
			require.Equal(t, uint(7), input.ProjectID)
			require.Equal(t, "weekly", input.Name)
			require.NotNil(t, input.Enabled)
			require.False(t, *input.Enabled)
			return &service.ProjectAnalysisPlan{ID: 11, ProjectID: 7, Name: "weekly"}, nil
		},
		get: func(ctx context.Context, id uint) (*service.ProjectAnalysisPlan, error) {
			require.Equal(t, uint(11), id)
			return &service.ProjectAnalysisPlan{ID: 11, ProjectID: 7, Name: "weekly"}, nil
		},
		search: func(ctx context.Context, query service.ProjectAnalysisPlanSearchQuery) (*service.ProjectAnalysisPlanPage, error) {
			require.Equal(t, uint(7), query.ProjectID)
			require.Equal(t, "weekly", query.Keyword)
			require.NotNil(t, query.Enabled)
			require.False(t, *query.Enabled)
			require.Equal(t, 2, query.Page)
			require.Equal(t, 10, query.Size)
			return &service.ProjectAnalysisPlanPage{
				Items: []service.ProjectAnalysisPlan{{ID: 11, Name: "weekly"}},
				Total: 1,
				Page:  2,
				Size:  10,
			}, nil
		},
		delete: func(ctx context.Context, ids []uint) error {
			require.Equal(t, []uint{11}, ids)
			return nil
		},
	})
	r.POST("/create", handler.Create)
	r.GET("/get", handler.Get)
	r.GET("/search", handler.Search)
	r.POST("/delete", handler.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(`{"projectId":7,"name":"weekly","enabled":false}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/get?id=11", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/search?projectId=7&keyword=weekly&enabled=false&page=2&size=10", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["total"])

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/delete", bytes.NewBufferString(`{"ids":[11]}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestProjectAnalysisPlanHandlerMapsErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewProjectAnalysisPlanHandler(&fakeProjectAnalysisPlanService{
		get: func(ctx context.Context, id uint) (*service.ProjectAnalysisPlan, error) {
			return nil, service.ErrProjectAnalysisPlanNotFound
		},
	})
	r.GET("/get", handler.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/get?id=404", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

type fakeProjectAnalysisPlanService struct {
	create func(context.Context, service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error)
	update func(context.Context, uint, service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error)
	get    func(context.Context, uint) (*service.ProjectAnalysisPlan, error)
	delete func(context.Context, []uint) error
	search func(context.Context, service.ProjectAnalysisPlanSearchQuery) (*service.ProjectAnalysisPlanPage, error)
}

func (s *fakeProjectAnalysisPlanService) Create(ctx context.Context, input service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error) {
	return s.create(ctx, input)
}

func (s *fakeProjectAnalysisPlanService) Update(ctx context.Context, id uint, input service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error) {
	return s.update(ctx, id, input)
}

func (s *fakeProjectAnalysisPlanService) Get(ctx context.Context, id uint) (*service.ProjectAnalysisPlan, error) {
	return s.get(ctx, id)
}

func (s *fakeProjectAnalysisPlanService) Delete(ctx context.Context, ids []uint) error {
	return s.delete(ctx, ids)
}

func (s *fakeProjectAnalysisPlanService) Search(ctx context.Context, query service.ProjectAnalysisPlanSearchQuery) (*service.ProjectAnalysisPlanPage, error) {
	return s.search(ctx, query)
}
