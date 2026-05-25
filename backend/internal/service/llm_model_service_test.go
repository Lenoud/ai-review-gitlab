package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLLMModelServiceCreateAndDefault(t *testing.T) {
	repo := newMemoryLLMModelRepository()
	svc := NewLLMModelService(repo, &fakeLLMConnectionChecker{})

	model, err := svc.Create(context.Background(), LLMModelInput{
		Provider:   "openai",
		ModelCode:  "gpt-4o-mini",
		APIBaseURL: "https://api.example.com/v1",
		APIKey:     "secret",
		MaxTokens:  4096,
		IsDefault:  true,
	})

	require.NoError(t, err)
	require.NotZero(t, model.ID)
	require.Equal(t, "openai", model.Provider)
	require.Equal(t, "gpt-4o-mini", model.ModelCode)
	require.True(t, model.IsDefault)

	defaultModel, err := svc.Default(context.Background())
	require.NoError(t, err)
	require.Equal(t, model.ID, defaultModel.ID)
}

func TestLLMModelServiceUpdateMissingModel(t *testing.T) {
	svc := NewLLMModelService(newMemoryLLMModelRepository(), &fakeLLMConnectionChecker{})

	_, err := svc.Update(context.Background(), 99, LLMModelInput{
		Provider:   "openai",
		ModelCode:  "gpt-4o-mini",
		APIBaseURL: "https://api.example.com/v1",
		APIKey:     "secret",
	})

	require.ErrorIs(t, err, ErrLLMModelNotFound)
}

func TestLLMModelServiceRejectsEmptyDeleteIDs(t *testing.T) {
	svc := NewLLMModelService(newMemoryLLMModelRepository(), &fakeLLMConnectionChecker{})

	err := svc.Delete(context.Background(), nil)

	require.ErrorIs(t, err, ErrInvalidLLMModelInput)
}

func TestLLMModelServiceSearchAndSetDefault(t *testing.T) {
	repo := newMemoryLLMModelRepository(
		&LLMModel{ID: 1, Provider: "openai", ModelCode: "gpt-4o-mini", APIBaseURL: "https://api.example.com/v1", APIKey: "a"},
		&LLMModel{ID: 2, Provider: "dashscope", ModelCode: "qwen-plus", APIBaseURL: "https://dashscope.example.com/compatible-mode/v1", APIKey: "b"},
	)
	svc := NewLLMModelService(repo, &fakeLLMConnectionChecker{})

	page, err := svc.Search(context.Background(), LLMModelSearchQuery{Keyword: "qwen", Page: 1, Size: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "qwen-plus", page.Items[0].ModelCode)

	require.NoError(t, svc.SetDefault(context.Background(), 2))
	defaultModel, err := svc.Default(context.Background())
	require.NoError(t, err)
	require.Equal(t, uint(2), defaultModel.ID)
}

func TestLLMModelServiceTestConnectionUsesChecker(t *testing.T) {
	checker := &fakeLLMConnectionChecker{}
	svc := NewLLMModelService(newMemoryLLMModelRepository(), checker)

	err := svc.TestConnection(context.Background(), LLMConnectionInput{
		Provider:   "openai",
		ModelCode:  "gpt-4o-mini",
		APIBaseURL: "https://api.example.com/v1",
		APIKey:     "secret",
	})

	require.NoError(t, err)
	require.Equal(t, "gpt-4o-mini", checker.last.ModelCode)
}

type memoryLLMModelRepository struct {
	models map[uint]*LLMModel
	nextID uint
}

func newMemoryLLMModelRepository(models ...*LLMModel) *memoryLLMModelRepository {
	repo := &memoryLLMModelRepository{models: map[uint]*LLMModel{}, nextID: 1}
	for _, model := range models {
		copy := *model
		repo.models[copy.ID] = &copy
		if copy.ID >= repo.nextID {
			repo.nextID = copy.ID + 1
		}
	}
	return repo
}

func (r *memoryLLMModelRepository) Create(ctx context.Context, input LLMModelInput) (*LLMModel, error) {
	model := llmModelFromInput(input)
	model.ID = r.nextID
	r.nextID++
	if model.IsDefault {
		r.clearDefault()
	}
	r.models[model.ID] = model
	return cloneLLMModel(model), nil
}

func (r *memoryLLMModelRepository) Update(ctx context.Context, id uint, input LLMModelInput) (*LLMModel, error) {
	if _, ok := r.models[id]; !ok {
		return nil, ErrLLMModelNotFound
	}
	model := llmModelFromInput(input)
	model.ID = id
	if model.IsDefault {
		r.clearDefault()
	}
	r.models[id] = model
	return cloneLLMModel(model), nil
}

func (r *memoryLLMModelRepository) FindByID(ctx context.Context, id uint) (*LLMModel, error) {
	model, ok := r.models[id]
	if !ok {
		return nil, ErrLLMModelNotFound
	}
	return cloneLLMModel(model), nil
}

func (r *memoryLLMModelRepository) Delete(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		delete(r.models, id)
	}
	return nil
}

func (r *memoryLLMModelRepository) Search(ctx context.Context, query LLMModelSearchQuery) (*LLMModelPage, error) {
	items := make([]LLMModel, 0)
	for _, model := range r.models {
		if query.Keyword == "" || containsFold(model.ModelCode, query.Keyword) || containsFold(model.Provider, query.Keyword) {
			items = append(items, *cloneLLMModel(model))
		}
	}
	return &LLMModelPage{Items: items, Total: int64(len(items)), Page: query.Page, Size: query.Size}, nil
}

func (r *memoryLLMModelRepository) Default(ctx context.Context) (*LLMModel, error) {
	for _, model := range r.models {
		if model.IsDefault {
			return cloneLLMModel(model), nil
		}
	}
	return nil, ErrLLMModelNotFound
}

func (r *memoryLLMModelRepository) SetDefault(ctx context.Context, id uint) error {
	if _, ok := r.models[id]; !ok {
		return ErrLLMModelNotFound
	}
	r.clearDefault()
	r.models[id].IsDefault = true
	return nil
}

func (r *memoryLLMModelRepository) clearDefault() {
	for _, model := range r.models {
		model.IsDefault = false
	}
}

type fakeLLMConnectionChecker struct {
	last LLMConnectionInput
	err  error
}

func (c *fakeLLMConnectionChecker) Check(ctx context.Context, input LLMConnectionInput) error {
	c.last = input
	return c.err
}

func TestFakeLLMConnectionCheckerPropagatesError(t *testing.T) {
	err := errors.New("failed")
	checker := &fakeLLMConnectionChecker{err: err}
	require.ErrorIs(t, checker.Check(context.Background(), LLMConnectionInput{}), err)
}
