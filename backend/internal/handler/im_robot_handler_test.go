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

func TestIMRobotHandlerCreateGetSearchListEnabledAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewIMRobotHandler(&fakeIMRobotService{
		create: func(ctx context.Context, input service.IMRobotInput) (*service.IMRobot, error) {
			require.Equal(t, service.IMRobotPlatformDingTalk, input.Platform)
			require.Equal(t, "alerts", input.Name)
			require.Equal(t, "https://example.com/webhook", input.WebhookURL)
			require.True(t, input.EnabledSet)
			require.False(t, input.Enabled)
			return &service.IMRobot{ID: 11, Name: input.Name, Enabled: input.Enabled}, nil
		},
		get: func(ctx context.Context, id uint) (*service.IMRobot, error) {
			require.Equal(t, uint(11), id)
			return &service.IMRobot{ID: 11, Name: "alerts", Enabled: true}, nil
		},
		search: func(ctx context.Context, query service.IMRobotSearchQuery) (*service.IMRobotPage, error) {
			require.Equal(t, "alerts", query.Keyword)
			require.Equal(t, service.IMRobotPlatformDingTalk, query.Platform)
			require.NotNil(t, query.Enabled)
			require.False(t, *query.Enabled)
			require.Equal(t, 2, query.Page)
			require.Equal(t, 5, query.Size)
			return &service.IMRobotPage{Items: []service.IMRobot{{ID: 11, Name: "alerts"}}, Total: 1, Page: 2, Size: 5}, nil
		},
		listEnabled: func(ctx context.Context) ([]service.IMRobot, error) {
			return []service.IMRobot{{ID: 11, Name: "alerts", Enabled: true}}, nil
		},
		testWebhook: func(ctx context.Context, input service.IMRobotTestWebhookInput) (*service.IMRobotTestWebhookResult, error) {
			require.Equal(t, service.IMRobotPlatformFeishu, input.Platform)
			require.Equal(t, "https://example.com/webhook", input.WebhookURL)
			return &service.IMRobotTestWebhookResult{Success: true, Message: "ok"}, nil
		},
		delete: func(ctx context.Context, ids []uint) error {
			require.Equal(t, []uint{11}, ids)
			return nil
		},
	})
	r.POST("/create", handler.Create)
	r.GET("/get", handler.Get)
	r.GET("/search", handler.Search)
	r.GET("/list-enabled", handler.ListEnabled)
	r.POST("/test-webhook", handler.TestWebhook)
	r.POST("/delete", handler.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(`{"platform":"dingtalk","name":"alerts","webhookUrl":"https://example.com/webhook","enabled":false}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/get?id=11", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/search?keyword=alerts&platform=dingtalk&enabled=false&page=2&size=5", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["total"])

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/list-enabled", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/test-webhook", bytes.NewBufferString(`{"platform":"feishu","webhookUrl":"https://example.com/webhook"}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data = body["data"].(map[string]any)
	require.Equal(t, true, data["success"])
	require.Equal(t, "ok", data["message"])

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/delete", bytes.NewBufferString(`{"ids":[11]}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestIMRobotHandlerMapsErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewIMRobotHandler(&fakeIMRobotService{
		get: func(ctx context.Context, id uint) (*service.IMRobot, error) {
			return nil, service.ErrIMRobotNotFound
		},
	})
	r.GET("/get", handler.Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/get?id=404", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestIMRobotHandlerMapsRobotInUse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewIMRobotHandler(&fakeIMRobotService{
		delete: func(ctx context.Context, ids []uint) error {
			return service.ErrIMRobotInUse
		},
	})
	r.POST("/delete", handler.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewBufferString(`{"ids":[7]}`))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

type fakeIMRobotService struct {
	create      func(context.Context, service.IMRobotInput) (*service.IMRobot, error)
	update      func(context.Context, uint, service.IMRobotInput) (*service.IMRobot, error)
	get         func(context.Context, uint) (*service.IMRobot, error)
	delete      func(context.Context, []uint) error
	search      func(context.Context, service.IMRobotSearchQuery) (*service.IMRobotPage, error)
	listEnabled func(context.Context) ([]service.IMRobot, error)
	testWebhook func(context.Context, service.IMRobotTestWebhookInput) (*service.IMRobotTestWebhookResult, error)
}

func (s *fakeIMRobotService) Create(ctx context.Context, input service.IMRobotInput) (*service.IMRobot, error) {
	return s.create(ctx, input)
}

func (s *fakeIMRobotService) Update(ctx context.Context, id uint, input service.IMRobotInput) (*service.IMRobot, error) {
	return s.update(ctx, id, input)
}

func (s *fakeIMRobotService) Get(ctx context.Context, id uint) (*service.IMRobot, error) {
	return s.get(ctx, id)
}

func (s *fakeIMRobotService) Delete(ctx context.Context, ids []uint) error {
	return s.delete(ctx, ids)
}

func (s *fakeIMRobotService) Search(ctx context.Context, query service.IMRobotSearchQuery) (*service.IMRobotPage, error) {
	return s.search(ctx, query)
}

func (s *fakeIMRobotService) ListEnabled(ctx context.Context) ([]service.IMRobot, error) {
	return s.listEnabled(ctx)
}

func (s *fakeIMRobotService) TestWebhook(ctx context.Context, input service.IMRobotTestWebhookInput) (*service.IMRobotTestWebhookResult, error) {
	return s.testWebhook(ctx, input)
}
