package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSysLogHandlerGetReturnsLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/sys-log/get", NewSysLogHandler(&fakeSysLogService{
		log: service.SysLog{ID: 7, Level: "INFO", Module: "REVIEW", Action: "send"},
	}).Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/sys-log/get?id=7", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(7), data["id"])
	require.Equal(t, "send", data["action"])
}

func TestSysLogHandlerGetMapsBadIDAndNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/bad", NewSysLogHandler(&fakeSysLogService{}).Get)
	r.GET("/missing", NewSysLogHandler(&fakeSysLogService{err: service.ErrSysLogNotFound}).Get)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bad?id=0", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/missing?id=9", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestSysLogHandlerSearchReturnsPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/sys-log/search", NewSysLogHandler(&fakeSysLogService{
		page: service.SysLogPage{Items: []service.SysLog{{ID: 1, Message: "sent"}}, Total: 1, Page: 1, Size: 20},
	}).Search)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/sys-log/search?level=INFO&module=REVIEW&action=send&message=sent&startTime=1000&endTime=2000", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["total"])
}

func TestSysLogHandlerSearchRejectsMalformedTimeFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin/sys-log/search", NewSysLogHandler(&fakeSysLogService{}).Search)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/sys-log/search?startTime=bad&endTime=2000", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/admin/sys-log/search?startTime=3000&endTime=2000", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

type fakeSysLogService struct {
	log  service.SysLog
	page service.SysLogPage
	err  error
}

func (s *fakeSysLogService) Get(ctx context.Context, id uint) (*service.SysLog, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.log, nil
}

func (s *fakeSysLogService) Search(ctx context.Context, query service.SysLogSearchQuery) (*service.SysLogPage, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &s.page, nil
}
