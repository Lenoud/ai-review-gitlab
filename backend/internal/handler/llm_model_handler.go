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

type LLMModelService interface {
	Create(ctx context.Context, input service.LLMModelInput) (*service.LLMModel, error)
	Update(ctx context.Context, id uint, input service.LLMModelInput) (*service.LLMModel, error)
	Get(ctx context.Context, id uint) (*service.LLMModel, error)
	Delete(ctx context.Context, ids []uint) error
	Search(ctx context.Context, query service.LLMModelSearchQuery) (*service.LLMModelPage, error)
	Default(ctx context.Context) (*service.LLMModel, error)
	SetDefault(ctx context.Context, id uint) error
	TestConnection(ctx context.Context, input service.LLMConnectionInput) error
}

type LLMModelHandler struct {
	models LLMModelService
}

func NewLLMModelHandler(models LLMModelService) *LLMModelHandler {
	return &LLMModelHandler{models: models}
}

func (h *LLMModelHandler) Create(c *gin.Context) {
	var req llmModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "模型参数错误")
		return
	}
	model, err := h.models.Create(c.Request.Context(), req.toInput())
	if err != nil {
		writeLLMModelError(c, err)
		return
	}
	response.Success(c, model)
}

func (h *LLMModelHandler) Update(c *gin.Context) {
	var req llmModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "模型参数错误")
		return
	}
	model, err := h.models.Update(c.Request.Context(), req.ID, req.toInput())
	if err != nil {
		writeLLMModelError(c, err)
		return
	}
	response.Success(c, model)
}

func (h *LLMModelHandler) Get(c *gin.Context) {
	id, ok := parseUintQuery(c, "id")
	if !ok {
		response.BadRequest(c, "模型ID不能为空")
		return
	}
	model, err := h.models.Get(c.Request.Context(), id)
	if err != nil {
		writeLLMModelError(c, err)
		return
	}
	response.Success(c, model)
}

func (h *LLMModelHandler) Delete(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择要删除的模型")
		return
	}
	if err := h.models.Delete(c.Request.Context(), req.IDs); err != nil {
		writeLLMModelError(c, err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

func (h *LLMModelHandler) Search(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := h.models.Search(c.Request.Context(), service.LLMModelSearchQuery{
		Keyword:  c.Query("keyword"),
		Provider: c.Query("provider"),
		Page:     page,
		Size:     size,
	})
	if err != nil {
		writeLLMModelError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *LLMModelHandler) Default(c *gin.Context) {
	model, err := h.models.Default(c.Request.Context())
	if err != nil {
		writeLLMModelError(c, err)
		return
	}
	response.Success(c, model)
}

func (h *LLMModelHandler) SetDefault(c *gin.Context) {
	var req setDefaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "模型ID不能为空")
		return
	}
	if err := h.models.SetDefault(c.Request.Context(), req.ID); err != nil {
		writeLLMModelError(c, err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func (h *LLMModelHandler) TestConnection(c *gin.Context) {
	var req llmConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "模型参数错误")
		return
	}
	if err := h.models.TestConnection(c.Request.Context(), req.toInput()); err != nil {
		writeLLMModelError(c, err)
		return
	}
	response.Success(c, gin.H{"ok": true})
}

func writeLLMModelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidLLMModelInput):
		response.BadRequest(c, "模型参数错误")
	case errors.Is(err, service.ErrLLMModelNotFound):
		response.Error(c, http.StatusNotFound, 40400, "模型不存在")
	default:
		response.Error(c, http.StatusInternalServerError, 50000, "模型操作失败")
	}
}

type llmModelRequest struct {
	ID         uint   `json:"id"`
	Provider   string `json:"provider"`
	ModelCode  string `json:"modelCode"`
	APIBaseURL string `json:"apiBaseUrl"`
	APIKey     string `json:"apiKey"`
	MaxTokens  int    `json:"maxTokens"`
	IsDefault  bool   `json:"isDefault"`
}

func (r llmModelRequest) toInput() service.LLMModelInput {
	return service.LLMModelInput{
		Provider:   r.Provider,
		ModelCode:  r.ModelCode,
		APIBaseURL: r.APIBaseURL,
		APIKey:     r.APIKey,
		MaxTokens:  r.MaxTokens,
		IsDefault:  r.IsDefault,
	}
}

type llmConnectionRequest struct {
	Provider   string `json:"provider"`
	ModelCode  string `json:"modelCode"`
	APIBaseURL string `json:"apiBaseUrl"`
	APIKey     string `json:"apiKey"`
	MaxTokens  int    `json:"maxTokens"`
}

func (r llmConnectionRequest) toInput() service.LLMConnectionInput {
	return service.LLMConnectionInput{
		Provider:   r.Provider,
		ModelCode:  r.ModelCode,
		APIBaseURL: r.APIBaseURL,
		APIKey:     r.APIKey,
		MaxTokens:  r.MaxTokens,
	}
}

type setDefaultRequest struct {
	ID uint `json:"id"`
}
