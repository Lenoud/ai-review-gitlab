package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestJWTAuthRejectsMissingBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", JWTAuth(&fakeTokenValidator{}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthAllowsValidAccessTokenAndSetsSubject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	validator := &fakeTokenValidator{
		subject: &service.AuthSubject{
			UserID:   1,
			Username: "admin",
			Nickname: "Administrator",
		},
	}
	r := gin.New()
	r.GET("/admin", JWTAuth(validator), func(c *gin.Context) {
		subject, ok := CurrentSubject(c)
		require.True(t, ok)
		require.Equal(t, uint(1), subject.UserID)
		require.Equal(t, "admin", subject.Username)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "access-token", validator.token)
}

func TestJWTAuthRejectsInvalidAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", JWTAuth(&fakeTokenValidator{err: service.ErrInvalidToken}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequirePermissionAllowsAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		SetCurrentSubject(c, service.AuthSubject{UserID: 1, Username: "admin", Roles: []string{"admin"}})
		c.Next()
	}, RequirePermission("project:write"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermissionAllowsExplicitPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		SetCurrentSubject(c, service.AuthSubject{UserID: 2, Username: "reviewer", Permissions: []string{"project:write"}})
		c.Next()
	}, RequirePermission("project:write"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermissionRejectsMissingPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		SetCurrentSubject(c, service.AuthSubject{UserID: 2, Username: "reviewer", Permissions: []string{"review-log:read"}})
		c.Next()
	}, RequirePermission("project:write"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

type fakeTokenValidator struct {
	token   string
	subject *service.AuthSubject
	err     error
}

func (v *fakeTokenValidator) ValidateAccessToken(ctx context.Context, token string) (*service.AuthSubject, error) {
	v.token = token
	if v.err != nil {
		return nil, v.err
	}
	return v.subject, nil
}
