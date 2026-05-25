package handler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ReviewTaskService interface {
	EnqueueGitLabWebhook(ctx context.Context, input service.GitLabWebhookInput) (*service.ReviewTaskEnqueueResult, error)
}

type WebhookHandler struct {
	tasks ReviewTaskService
}

func NewWebhookHandler(tasks ReviewTaskService) *WebhookHandler {
	return &WebhookHandler{tasks: tasks}
}

func (h *WebhookHandler) GitLab(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Webhook payload 无效")
		return
	}
	result, err := h.tasks.EnqueueGitLabWebhook(c.Request.Context(), service.GitLabWebhookInput{
		Event:   c.GetHeader("X-Gitlab-Event"),
		Payload: payload,
	})
	if err != nil {
		writeWebhookError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, response.Envelope{
		Code:    0,
		Message: "success",
		Data:    result,
	})
}

func writeWebhookError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidReviewTaskInput):
		response.BadRequest(c, "Webhook payload 无效")
	case errors.Is(err, service.ErrUnsupportedWebhookEvent):
		response.BadRequest(c, "不支持的 Webhook 事件")
	case errors.Is(err, service.ErrReviewTaskProjectNotFound):
		response.Error(c, http.StatusNotFound, 40400, "项目不存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "Webhook 处理失败")
	}
}
