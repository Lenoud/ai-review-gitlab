package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type IMRobotService interface {
	Create(ctx context.Context, input service.IMRobotInput) (*service.IMRobot, error)
	Update(ctx context.Context, id uint, input service.IMRobotInput) (*service.IMRobot, error)
	Get(ctx context.Context, id uint) (*service.IMRobot, error)
	Delete(ctx context.Context, ids []uint) error
	Search(ctx context.Context, query service.IMRobotSearchQuery) (*service.IMRobotPage, error)
	ListEnabled(ctx context.Context) ([]service.IMRobot, error)
	TestWebhook(ctx context.Context, input service.IMRobotTestWebhookInput) (*service.IMRobotTestWebhookResult, error)
}

type IMRobotHandler struct {
	robots IMRobotService
}

func NewIMRobotHandler(robots IMRobotService) *IMRobotHandler {
	return &IMRobotHandler{robots: robots}
}

func (h *IMRobotHandler) Create(c *gin.Context) {
	var req imRobotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "IM机器人参数错误")
		return
	}
	robot, err := h.robots.Create(c.Request.Context(), req.toInput())
	if err != nil {
		writeIMRobotError(c, err)
		return
	}
	response.Success(c, robot)
}

func (h *IMRobotHandler) Update(c *gin.Context) {
	var req imRobotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "IM机器人参数错误")
		return
	}
	robot, err := h.robots.Update(c.Request.Context(), req.ID, req.toInput())
	if err != nil {
		writeIMRobotError(c, err)
		return
	}
	response.Success(c, robot)
}

func (h *IMRobotHandler) Get(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "IM机器人ID不能为空")
		return
	}
	robot, err := h.robots.Get(c.Request.Context(), id)
	if err != nil {
		writeIMRobotError(c, err)
		return
	}
	response.Success(c, robot)
}

func (h *IMRobotHandler) Delete(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择要删除的IM机器人")
		return
	}
	if err := h.robots.Delete(c.Request.Context(), req.IDs); err != nil {
		writeIMRobotError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *IMRobotHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := h.robots.Search(c.Request.Context(), service.IMRobotSearchQuery{
		Keyword:  c.Query("keyword"),
		Platform: c.Query("platform"),
		Enabled:  parseOptionalBoolPtr(c.Query("enabled")),
		Page:     page,
		Size:     size,
	})
	if err != nil {
		writeIMRobotError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *IMRobotHandler) ListEnabled(c *gin.Context) {
	result, err := h.robots.ListEnabled(c.Request.Context())
	if err != nil {
		writeIMRobotError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *IMRobotHandler) TestWebhook(c *gin.Context) {
	var req imRobotTestWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "IM机器人Webhook参数错误")
		return
	}
	result, err := h.robots.TestWebhook(c.Request.Context(), service.IMRobotTestWebhookInput{
		Platform:   req.Platform,
		WebhookURL: req.WebhookURL,
	})
	if err != nil {
		writeIMRobotError(c, err)
		return
	}
	response.Success(c, result)
}

type imRobotRequest struct {
	ID         uint   `json:"id"`
	Platform   string `json:"platform"`
	Name       string `json:"name"`
	WebhookURL string `json:"webhookUrl"`
	Secret     string `json:"secret"`
	Enabled    *bool  `json:"enabled"`
}

func (r imRobotRequest) toInput() service.IMRobotInput {
	input := service.IMRobotInput{
		Platform:   r.Platform,
		Name:       r.Name,
		WebhookURL: r.WebhookURL,
		Secret:     r.Secret,
	}
	if r.Enabled != nil {
		input.EnabledSet = true
		input.Enabled = *r.Enabled
	}
	return input
}

type imRobotTestWebhookRequest struct {
	Platform   string `json:"platform"`
	WebhookURL string `json:"webhookUrl"`
}

func writeIMRobotError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidIMRobotInput):
		response.BadRequest(c, "IM机器人参数错误")
	case errors.Is(err, service.ErrIMRobotNotFound):
		response.Error(c, http.StatusNotFound, 40400, "IM机器人不存在")
	case errors.Is(err, service.ErrIMRobotInUse):
		response.Error(c, http.StatusConflict, 40900, "IM机器人已被项目或分析计划使用")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "IM机器人操作失败")
	}
}
