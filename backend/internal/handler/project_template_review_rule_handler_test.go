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

func TestProjectTemplateReviewRuleHandlerCreateUpdateGetListAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewProjectTemplateReviewRuleHandler(&fakeProjectTemplateReviewRuleService{
		create: func(ctx context.Context, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error) {
			require.Equal(t, uint(7), input.TemplateID)
			require.Equal(t, "Controller rules", input.Name)
			require.Equal(t, []string{"*.go"}, input.GlobPatterns)
			require.False(t, input.Enabled)
			require.True(t, input.EnabledIsSet)
			return &service.ProjectTemplateReviewRule{ID: 11, TemplateID: 7, Name: input.Name, Enabled: input.Enabled}, nil
		},
		update: func(ctx context.Context, id uint, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error) {
			require.Equal(t, uint(11), id)
			require.Equal(t, uint(7), input.TemplateID)
			return &service.ProjectTemplateReviewRule{ID: 11, TemplateID: 7, Name: input.Name, Enabled: true}, nil
		},
		get: func(ctx context.Context, id uint) (*service.ProjectTemplateReviewRule, error) {
			require.Equal(t, uint(11), id)
			return &service.ProjectTemplateReviewRule{ID: 11, TemplateID: 7, Name: "Controller rules"}, nil
		},
		listByTemplateID: func(ctx context.Context, templateID uint) ([]service.ProjectTemplateReviewRule, error) {
			require.Equal(t, uint(7), templateID)
			return []service.ProjectTemplateReviewRule{{ID: 11, TemplateID: 7, Name: "Controller rules"}}, nil
		},
		delete: func(ctx context.Context, id uint) error {
			require.Equal(t, uint(11), id)
			return nil
		},
	})
	r.POST("/create", handler.Create)
	r.POST("/update", handler.Update)
	r.GET("/get", handler.Get)
	r.GET("/list-by-template-id", handler.ListByTemplateID)
	r.POST("/delete", handler.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(`{"projectTemplateId":7,"name":"Controller rules","globPatterns":["*.go"],"content":"Use context","enabled":false}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(`{"id":11,"projectTemplateId":7,"name":"Controller rules","globPatterns":["*.go"],"content":"Use context"}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/get?id=11", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/list-by-template-id?projectTemplateId=7", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].([]any)
	require.Len(t, data, 1)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/delete", bytes.NewBufferString(`{"id":11}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestProjectTemplateReviewRuleHandlerMapsErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewProjectTemplateReviewRuleHandler(&fakeProjectTemplateReviewRuleService{
		get: func(ctx context.Context, id uint) (*service.ProjectTemplateReviewRule, error) {
			return nil, service.ErrProjectTemplateReviewRuleNotFound
		},
	})
	r.GET("/get", handler.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/get?id=404", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

type fakeProjectTemplateReviewRuleService struct {
	create           func(context.Context, service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error)
	update           func(context.Context, uint, service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error)
	get              func(context.Context, uint) (*service.ProjectTemplateReviewRule, error)
	delete           func(context.Context, uint) error
	listByTemplateID func(context.Context, uint) ([]service.ProjectTemplateReviewRule, error)
}

func (s *fakeProjectTemplateReviewRuleService) Create(ctx context.Context, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error) {
	return s.create(ctx, input)
}

func (s *fakeProjectTemplateReviewRuleService) Update(ctx context.Context, id uint, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error) {
	return s.update(ctx, id, input)
}

func (s *fakeProjectTemplateReviewRuleService) Get(ctx context.Context, id uint) (*service.ProjectTemplateReviewRule, error) {
	return s.get(ctx, id)
}

func (s *fakeProjectTemplateReviewRuleService) Delete(ctx context.Context, id uint) error {
	return s.delete(ctx, id)
}

func (s *fakeProjectTemplateReviewRuleService) ListByTemplateID(ctx context.Context, templateID uint) ([]service.ProjectTemplateReviewRule, error) {
	return s.listByTemplateID(ctx, templateID)
}
