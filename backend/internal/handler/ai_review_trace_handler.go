package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/pkg/response"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AIReviewTraceService interface {
	Create(ctx context.Context, input service.AIReviewTraceInput) (*service.AIReviewTrace, error)
	Get(ctx context.Context, eventType string, eventID uint) (*service.AIReviewTrace, error)
}

type AIReviewTraceHandler struct {
	traces AIReviewTraceService
}

func NewAIReviewTraceHandler(traces AIReviewTraceService) *AIReviewTraceHandler {
	return &AIReviewTraceHandler{traces: traces}
}

func (h *AIReviewTraceHandler) Create(c *gin.Context) {
	var req aiReviewTraceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "AI评审追踪参数错误")
		return
	}
	trace, err := h.traces.Create(c.Request.Context(), req.toInput())
	if err != nil {
		writeAIReviewTraceError(c, err)
		return
	}
	response.Success(c, trace)
}

func (h *AIReviewTraceHandler) Get(c *gin.Context) {
	eventID, ok := parseUintQuery(c, "reviewEventId")
	if !ok {
		response.BadRequest(c, "AI评审追踪参数错误")
		return
	}
	trace, err := h.traces.Get(c.Request.Context(), c.Query("reviewEventType"), eventID)
	if err != nil {
		writeAIReviewTraceError(c, err)
		return
	}
	response.Success(c, trace)
}

func writeAIReviewTraceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidAIReviewTraceInput):
		response.BadRequest(c, "AI评审追踪参数错误")
	case errors.Is(err, service.ErrAIReviewTraceNotFound):
		response.Error(c, http.StatusNotFound, 40400, "AI评审追踪不存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "AI评审追踪操作失败")
	}
}

type aiReviewTraceRequest struct {
	ReviewEventType string `json:"reviewEventType"`
	ReviewEventID   uint   `json:"reviewEventId"`
	Prompt          string `json:"prompt"`
	Response        string `json:"response"`
	Provider        string `json:"provider"`
	ModelCode       string `json:"modelCode"`
}

func (r aiReviewTraceRequest) toInput() service.AIReviewTraceInput {
	return service.AIReviewTraceInput{
		ReviewEventType: r.ReviewEventType,
		ReviewEventID:   r.ReviewEventID,
		Prompt:          r.Prompt,
		Response:        r.Response,
		Provider:        r.Provider,
		ModelCode:       r.ModelCode,
	}
}
