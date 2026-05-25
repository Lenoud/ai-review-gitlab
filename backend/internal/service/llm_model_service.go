package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var (
	ErrLLMModelNotFound     = errors.New("llm model not found")
	ErrInvalidLLMModelInput = errors.New("invalid llm model input")
)

type LLMModel struct {
	ID         uint   `json:"id"`
	Provider   string `json:"provider"`
	ModelCode  string `json:"modelCode"`
	APIBaseURL string `json:"apiBaseUrl"`
	APIKey     string `json:"apiKey,omitempty"`
	MaxTokens  int    `json:"maxTokens"`
	IsDefault  bool   `json:"isDefault"`
}

type LLMModelInput struct {
	Provider   string
	ModelCode  string
	APIBaseURL string
	APIKey     string
	MaxTokens  int
	IsDefault  bool
}

type LLMConnectionInput struct {
	Provider   string
	ModelCode  string
	APIBaseURL string
	APIKey     string
	MaxTokens  int
}

type LLMModelSearchQuery struct {
	Keyword  string
	Provider string
	Page     int
	Size     int
}

type LLMModelPage struct {
	Items []LLMModel `json:"items"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

type LLMModelRepository interface {
	Create(ctx context.Context, input LLMModelInput) (*LLMModel, error)
	Update(ctx context.Context, id uint, input LLMModelInput) (*LLMModel, error)
	FindByID(ctx context.Context, id uint) (*LLMModel, error)
	Delete(ctx context.Context, ids []uint) error
	Search(ctx context.Context, query LLMModelSearchQuery) (*LLMModelPage, error)
	Default(ctx context.Context) (*LLMModel, error)
	SetDefault(ctx context.Context, id uint) error
}

type LLMConnectionChecker interface {
	Check(ctx context.Context, input LLMConnectionInput) error
}

type LLMModelService struct {
	models  LLMModelRepository
	checker LLMConnectionChecker
}

func NewLLMModelService(models LLMModelRepository, checker LLMConnectionChecker) *LLMModelService {
	return &LLMModelService{models: models, checker: checker}
}

func (s *LLMModelService) Create(ctx context.Context, input LLMModelInput) (*LLMModel, error) {
	normalized, err := normalizeLLMModelInput(input)
	if err != nil {
		return nil, err
	}
	return s.models.Create(ctx, normalized)
}

func (s *LLMModelService) Update(ctx context.Context, id uint, input LLMModelInput) (*LLMModel, error) {
	if id == 0 {
		return nil, ErrInvalidLLMModelInput
	}
	if _, err := s.models.FindByID(ctx, id); err != nil {
		return nil, err
	}
	normalized, err := normalizeLLMModelInput(input)
	if err != nil {
		return nil, err
	}
	return s.models.Update(ctx, id, normalized)
}

func (s *LLMModelService) Get(ctx context.Context, id uint) (*LLMModel, error) {
	if id == 0 {
		return nil, ErrInvalidLLMModelInput
	}
	return s.models.FindByID(ctx, id)
}

func (s *LLMModelService) Delete(ctx context.Context, ids []uint) error {
	cleanIDs := make([]uint, 0, len(ids))
	seen := map[uint]struct{}{}
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		cleanIDs = append(cleanIDs, id)
	}
	if len(cleanIDs) == 0 {
		return ErrInvalidLLMModelInput
	}
	return s.models.Delete(ctx, cleanIDs)
}

func (s *LLMModelService) Search(ctx context.Context, query LLMModelSearchQuery) (*LLMModelPage, error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Provider = strings.TrimSpace(query.Provider)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 20
	}
	if query.Size > 200 {
		query.Size = 200
	}
	return s.models.Search(ctx, query)
}

func (s *LLMModelService) Default(ctx context.Context) (*LLMModel, error) {
	return s.models.Default(ctx)
}

func (s *LLMModelService) SetDefault(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrInvalidLLMModelInput
	}
	return s.models.SetDefault(ctx, id)
}

func (s *LLMModelService) TestConnection(ctx context.Context, input LLMConnectionInput) error {
	normalized, err := normalizeLLMConnectionInput(input)
	if err != nil {
		return err
	}
	if s.checker == nil {
		return ErrInvalidLLMModelInput
	}
	return s.checker.Check(ctx, normalized)
}

func normalizeLLMModelInput(input LLMModelInput) (LLMModelInput, error) {
	connection, err := normalizeLLMConnectionInput(LLMConnectionInput{
		Provider:   input.Provider,
		ModelCode:  input.ModelCode,
		APIBaseURL: input.APIBaseURL,
		APIKey:     input.APIKey,
		MaxTokens:  input.MaxTokens,
	})
	if err != nil {
		return LLMModelInput{}, err
	}
	return LLMModelInput{
		Provider:   connection.Provider,
		ModelCode:  connection.ModelCode,
		APIBaseURL: connection.APIBaseURL,
		APIKey:     connection.APIKey,
		MaxTokens:  connection.MaxTokens,
		IsDefault:  input.IsDefault,
	}, nil
}

func normalizeLLMConnectionInput(input LLMConnectionInput) (LLMConnectionInput, error) {
	input.Provider = strings.TrimSpace(input.Provider)
	input.ModelCode = strings.TrimSpace(input.ModelCode)
	input.APIBaseURL = strings.TrimRight(strings.TrimSpace(input.APIBaseURL), "/")
	input.APIKey = strings.TrimSpace(input.APIKey)
	if input.Provider == "" || input.ModelCode == "" || input.APIBaseURL == "" || input.APIKey == "" {
		return LLMConnectionInput{}, ErrInvalidLLMModelInput
	}
	if _, err := url.ParseRequestURI(input.APIBaseURL); err != nil {
		return LLMConnectionInput{}, ErrInvalidLLMModelInput
	}
	if input.MaxTokens <= 0 {
		input.MaxTokens = 4096
	}
	return input, nil
}

func llmModelFromInput(input LLMModelInput) *LLMModel {
	return &LLMModel{
		Provider:   input.Provider,
		ModelCode:  input.ModelCode,
		APIBaseURL: input.APIBaseURL,
		APIKey:     input.APIKey,
		MaxTokens:  input.MaxTokens,
		IsDefault:  input.IsDefault,
	}
}

func cloneLLMModel(model *LLMModel) *LLMModel {
	if model == nil {
		return nil
	}
	copy := *model
	return &copy
}
