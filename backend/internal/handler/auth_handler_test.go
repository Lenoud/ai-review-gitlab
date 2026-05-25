package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/middleware"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAuthHandlerLoginReturnsTokenPair(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authSvc := &fakeAuthService{
		login: func(ctx context.Context, input service.LoginInput) (*service.TokenPair, error) {
			require.Equal(t, "admin", input.Username)
			require.Equal(t, "secret", input.Password)
			return &service.TokenPair{
				AccessToken:      "access-token",
				RefreshToken:     "refresh-token",
				TokenType:        "Bearer",
				ExpiresIn:        1800,
				RefreshExpiresIn: 2592000,
			}, nil
		},
	}
	r := gin.New()
	r.POST("/login", NewAuthHandler(authSvc).Login)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, float64(0), body["code"])
	data := body["data"].(map[string]any)
	require.Equal(t, "access-token", data["accessToken"])
	require.Equal(t, "refresh-token", data["refreshToken"])
	require.Equal(t, "Bearer", data["tokenType"])
}

func TestAuthHandlerLoginRejectsMissingCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", NewAuthHandler(&fakeAuthService{}).Login)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"","password":""}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandlerMeReturnsCurrentSubject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me", func(c *gin.Context) {
		middleware.SetCurrentSubject(c, service.AuthSubject{
			UserID:   1,
			Username: "admin",
			Nickname: "Administrator",
		})
	}, NewAuthHandler(nil).Me)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	require.Equal(t, float64(1), data["id"])
	require.Equal(t, "admin", data["username"])
}

type fakeAuthService struct {
	login   func(context.Context, service.LoginInput) (*service.TokenPair, error)
	refresh func(context.Context, string) (*service.TokenPair, error)
}

func (s *fakeAuthService) Login(ctx context.Context, input service.LoginInput) (*service.TokenPair, error) {
	if s.login == nil {
		return nil, errors.New("unexpected login")
	}
	return s.login(ctx, input)
}

func (s *fakeAuthService) Refresh(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
	if s.refresh == nil {
		return nil, errors.New("unexpected refresh")
	}
	return s.refresh(ctx, refreshToken)
}
