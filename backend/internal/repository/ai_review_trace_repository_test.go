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

func TestAIReviewTraceRepositoryCreateAndFindByReviewEvent(t *testing.T) {
	db := openAIReviewTraceRepositoryTestDB(t)
	repo := NewAIReviewTraceRepository(db)

	trace, err := repo.Create(context.Background(), service.AIReviewTraceInput{
		ReviewEventType: "push",
		ReviewEventID:   101,
		Prompt:          "system: review\nuser: diff",
		Response:        "review ok",
		Provider:        "openai",
		ModelCode:       "gpt-test",
	})

	require.NoError(t, err)
	require.NotZero(t, trace.ID)
	require.Equal(t, "push", trace.ReviewEventType)
	require.Equal(t, uint(101), trace.ReviewEventID)
	require.NotZero(t, trace.CreatedAt)

	got, err := repo.FindByReviewEvent(context.Background(), "push", 101)
	require.NoError(t, err)
	require.Equal(t, trace.ID, got.ID)
	require.Equal(t, "system: review\nuser: diff", got.Prompt)
	require.Equal(t, "review ok", got.Response)
	require.Equal(t, "openai", got.Provider)
	require.Equal(t, "gpt-test", got.ModelCode)
}

func TestAIReviewTraceRepositoryFindByReviewEventNotFound(t *testing.T) {
	db := openAIReviewTraceRepositoryTestDB(t)
	repo := NewAIReviewTraceRepository(db)

	_, err := repo.FindByReviewEvent(context.Background(), "merge_request", 202)

	require.ErrorIs(t, err, service.ErrAIReviewTraceNotFound)
}

func openAIReviewTraceRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AIReviewTrace{}))
	return db
}
