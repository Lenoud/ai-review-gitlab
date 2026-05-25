package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/middleware"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthService interface {
	Login(ctx context.Context, input service.LoginInput) (*service.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*service.TokenPair, error)
}

type AuthHandler struct {
	auth AuthService
}

func NewAuthHandler(auth AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "用户名和密码不能为空")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		response.BadRequest(c, "用户名和密码不能为空")
		return
	}
	if h.auth == nil {
		response.Error(c, http.StatusInternalServerError, 50000, "认证服务未初始化")
		return
	}

	tokens, err := h.auth.Login(c.Request.Context(), service.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, tokens)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Refresh Token 不能为空")
		return
	}
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		response.BadRequest(c, "Refresh Token 不能为空")
		return
	}
	if h.auth == nil {
		response.Error(c, http.StatusInternalServerError, 50000, "认证服务未初始化")
		return
	}

	tokens, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	response.Success(c, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	response.Success(c, gin.H{"ok": true})
}

func (h *AuthHandler) Me(c *gin.Context) {
	subject, ok := middleware.CurrentSubject(c)
	if !ok {
		response.Unauthorized(c, "未认证")
		return
	}
	response.Success(c, subject)
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		response.Unauthorized(c, "用户名或密码错误")
	case errors.Is(err, service.ErrInvalidToken):
		response.Unauthorized(c, "Token 无效")
	case errors.Is(err, service.ErrUserDisabled):
		response.Forbidden(c, "用户已禁用")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "认证失败")
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}
