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

func TestSystemHandlerOpenInfoReturnsPublicConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/open/system/info", NewSystemHandler(&fakeSystemService{
		config: service.SystemConfig{Version: "2.0.0", SiteName: "Review Hub", SiteNotice: "maintenance", BaseURL: "https://cr.example.com"},
	}).OpenInfo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/open/system/info", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, "Review Hub", data["siteName"])
	require.Equal(t, "maintenance", data["siteNotice"])
	require.NotContains(t, data, "baseUrl")
}

func TestSystemHandlerAdminConfigInfoAndUpdateBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeSystemService{
		config:  service.SystemConfig{Version: "2.0.0", SiteName: "Review Hub", SiteNotice: "notice", BaseURL: "https://old.example.com"},
		updated: service.SystemConfig{Version: "2.0.0", SiteName: "Review Hub", SiteNotice: "notice", BaseURL: "https://cr.example.com"},
	}
	r := gin.New()
	h := NewSystemHandler(fake)
	r.GET("/admin/system/config", h.Config)
	r.GET("/admin/system/info", h.Info)
	r.POST("/admin/system/config/base-url", h.UpdateBaseURL)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/system/config", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, "https://old.example.com", data["baseUrl"])

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/system/info", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/admin/system/config/base-url", strings.NewReader(`{"baseUrl":"https://cr.example.com/"}`))
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "https://cr.example.com/", fake.updatedBaseURL)
}

func TestSystemHandlerUpdateBaseURLMapsValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/admin/system/config/base-url", NewSystemHandler(&fakeSystemService{err: service.ErrInvalidSystemConfigInput}).UpdateBaseURL)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/system/config/base-url", strings.NewReader(`{"baseUrl":"/relative"}`))
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

type fakeSystemService struct {
	config         service.SystemConfig
	updated        service.SystemConfig
	updatedBaseURL string
	err            error
}

func (s *fakeSystemService) GetConfig(ctx context.Context) (*service.SystemConfig, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.config, nil
}

func (s *fakeSystemService) UpdateBaseURL(ctx context.Context, baseURL string) (*service.SystemConfig, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.updatedBaseURL = baseURL
	return &s.updated, nil
}
