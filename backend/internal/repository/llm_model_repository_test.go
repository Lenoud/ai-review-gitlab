package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLLMModelRepositoryCRUDSearchAndDefault(t *testing.T) {
	db := openLLMModelRepositoryTestDB(t)
	repo := NewLLMModelRepository(db)

	first, err := repo.Create(context.Background(), service.LLMModelInput{
		Provider:   "openai",
		ModelCode:  "gpt-4o-mini",
		APIBaseURL: "https://api.example.com/v1",
		APIKey:     "a",
		MaxTokens:  4096,
		IsDefault:  true,
	})
	require.NoError(t, err)
	require.True(t, first.IsDefault)

	second, err := repo.Create(context.Background(), service.LLMModelInput{
		Provider:   "dashscope",
		ModelCode:  "qwen-plus",
		APIBaseURL: "https://dashscope.example.com/compatible-mode/v1",
		APIKey:     "b",
		MaxTokens:  8192,
	})
	require.NoError(t, err)

	require.NoError(t, repo.SetDefault(context.Background(), second.ID))
	defaultModel, err := repo.Default(context.Background())
	require.NoError(t, err)
	require.Equal(t, second.ID, defaultModel.ID)

	page, err := repo.Search(context.Background(), service.LLMModelSearchQuery{Keyword: "qwen", Page: 1, Size: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)

	updated, err := repo.Update(context.Background(), second.ID, service.LLMModelInput{
		Provider:   "dashscope",
		ModelCode:  "qwen-max",
		APIBaseURL: "https://dashscope.example.com/compatible-mode/v1",
		APIKey:     "b",
		MaxTokens:  8192,
		IsDefault:  true,
	})
	require.NoError(t, err)
	require.Equal(t, "qwen-max", updated.ModelCode)

	require.NoError(t, repo.Delete(context.Background(), []uint{first.ID}))
	_, err = repo.FindByID(context.Background(), first.ID)
	require.ErrorIs(t, err, service.ErrLLMModelNotFound)
}

func openLLMModelRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.LLMModel{}))
	return db
}
