package middleware

import (
	"context"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

const currentSubjectKey = "auth.subject"

type TokenValidator interface {
	ValidateAccessToken(ctx context.Context, token string) (*service.AuthSubject, error)
}

func JWTAuth(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			response.Unauthorized(c, "未认证")
			c.Abort()
			return
		}
		if validator == nil {
			response.Unauthorized(c, "Token 无效")
			c.Abort()
			return
		}
		subject, err := validator.ValidateAccessToken(c.Request.Context(), strings.TrimSpace(parts[1]))
		if err != nil {
			response.Unauthorized(c, "Token 无效")
			c.Abort()
			return
		}
		SetCurrentSubject(c, *subject)
		c.Next()
	}
}

func SetCurrentSubject(c *gin.Context, subject service.AuthSubject) {
	c.Set(currentSubjectKey, subject)
}

func CurrentSubject(c *gin.Context) (service.AuthSubject, bool) {
	value, ok := c.Get(currentSubjectKey)
	if !ok {
		return service.AuthSubject{}, false
	}
	subject, ok := value.(service.AuthSubject)
	return subject, ok
}

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permission = strings.TrimSpace(permission)
		if permission == "" {
			c.Next()
			return
		}
		subject, ok := CurrentSubject(c)
		if !ok {
			response.Unauthorized(c, "未认证")
			c.Abort()
			return
		}
		if !subjectAllows(subject, permission) {
			response.Forbidden(c, "无权限")
			c.Abort()
			return
		}
		c.Set("permission", permission)
		c.Next()
	}
}

func subjectAllows(subject service.AuthSubject, permission string) bool {
	for _, role := range subject.Roles {
		if strings.EqualFold(strings.TrimSpace(role), "admin") {
			return true
		}
	}
	for _, allowed := range subject.Permissions {
		if strings.EqualFold(strings.TrimSpace(allowed), permission) {
			return true
		}
	}
	return false
}
