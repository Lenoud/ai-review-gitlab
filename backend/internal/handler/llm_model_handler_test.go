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

func TestLLMModelHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	modelSvc := &fakeLLMModelService{
		create: func(ctx context.Context, input service.LLMModelInput) (*service.LLMModel, error) {
			require.Equal(t, "openai", input.Provider)
			require.Equal(t, "gpt-4o-mini", input.ModelCode)
			return &service.LLMModel{ID: 1, Provider: input.Provider, ModelCode: input.ModelCode, APIBaseURL: input.APIBaseURL}, nil
		},
	}
	r := gin.New()
	r.POST("/llm-model/create", NewLLMModelHandler(modelSvc).Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/llm-model/create", strings.NewReader(`{"provider":"openai","modelCode":"gpt-4o-mini","apiBaseUrl":"https://api.example.com/v1","apiKey":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["id"])
}

func TestLLMModelHandlerDefaultNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/llm-model/default", NewLLMModelHandler(&fakeLLMModelService{
		defaultModel: func(ctx context.Context) (*service.LLMModel, error) {
			return nil, service.ErrLLMModelNotFound
		},
	}).Default)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/llm-model/default", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestLLMModelHandlerTestConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/llm-test/connection", NewLLMModelHandler(&fakeLLMModelService{
		testConnection: func(ctx context.Context, input service.LLMConnectionInput) error {
			require.Equal(t, "gpt-4o-mini", input.ModelCode)
			return nil
		},
	}).TestConnection)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/llm-test/connection", strings.NewReader(`{"provider":"openai","modelCode":"gpt-4o-mini","apiBaseUrl":"https://api.example.com/v1","apiKey":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

type fakeLLMModelService struct {
	create         func(context.Context, service.LLMModelInput) (*service.LLMModel, error)
	update         func(context.Context, uint, service.LLMModelInput) (*service.LLMModel, error)
	get            func(context.Context, uint) (*service.LLMModel, error)
	delete         func(context.Context, []uint) error
	search         func(context.Context, service.LLMModelSearchQuery) (*service.LLMModelPage, error)
	defaultModel   func(context.Context) (*service.LLMModel, error)
	setDefault     func(context.Context, uint) error
	testConnection func(context.Context, service.LLMConnectionInput) error
}

func (s *fakeLLMModelService) Create(ctx context.Context, input service.LLMModelInput) (*service.LLMModel, error) {
	return s.create(ctx, input)
}
func (s *fakeLLMModelService) Update(ctx context.Context, id uint, input service.LLMModelInput) (*service.LLMModel, error) {
	return s.update(ctx, id, input)
}
func (s *fakeLLMModelService) Get(ctx context.Context, id uint) (*service.LLMModel, error) {
	return s.get(ctx, id)
}
func (s *fakeLLMModelService) Delete(ctx context.Context, ids []uint) error {
	return s.delete(ctx, ids)
}
func (s *fakeLLMModelService) Search(ctx context.Context, query service.LLMModelSearchQuery) (*service.LLMModelPage, error) {
	return s.search(ctx, query)
}
func (s *fakeLLMModelService) Default(ctx context.Context) (*service.LLMModel, error) {
	return s.defaultModel(ctx)
}
func (s *fakeLLMModelService) SetDefault(ctx context.Context, id uint) error {
	return s.setDefault(ctx, id)
}
func (s *fakeLLMModelService) TestConnection(ctx context.Context, input service.LLMConnectionInput) error {
	return s.testConnection(ctx, input)
}
