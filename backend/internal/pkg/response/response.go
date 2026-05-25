package response

import (
	"net/http"

	apperrors "github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/errors"
	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, code int, message string) {
	c.JSON(statusCode, Envelope{
		Code:    code,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, apperrors.CodeBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, apperrors.CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, apperrors.CodeForbidden, message)
}

func NotImplemented(c *gin.Context) {
	Error(c, http.StatusNotImplemented, apperrors.CodeNotImplemented, "接口暂未实现")
}
