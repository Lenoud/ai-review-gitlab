package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReviewTaskRepositoryCreateDedupesClaimAndTransitions(t *testing.T) {
	db := openReviewTaskRepositoryTestDB(t)
	repo := NewReviewTaskRepository(db)
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)

	task, duplicate, err := repo.CreateOrGetByDedupeKey(context.Background(), service.ReviewTaskCreateInput{
		ProjectID:   1,
		EventType:   service.ReviewTaskEventPush,
		DedupeKey:   "dedupe",
		PayloadJSON: []byte(`{"ok":true}`),
		MaxAttempts: 3,
		NextRunAt:   now,
	})
	require.NoError(t, err)
	require.False(t, duplicate)
	require.Equal(t, service.ReviewTaskStatusPending, task.Status)

	again, duplicate, err := repo.CreateOrGetByDedupeKey(context.Background(), service.ReviewTaskCreateInput{
		ProjectID:   1,
		EventType:   service.ReviewTaskEventPush,
		DedupeKey:   "dedupe",
		PayloadJSON: []byte(`{}`),
		MaxAttempts: 3,
		NextRunAt:   now,
	})
	require.NoError(t, err)
	require.True(t, duplicate)
	require.Equal(t, task.ID, again.ID)

	claimed, err := repo.ClaimNext(context.Background(), "worker-1", now)
	require.NoError(t, err)
	require.Equal(t, task.ID, claimed.ID)
	require.Equal(t, service.ReviewTaskStatusRunning, claimed.Status)

	attempt, err := repo.StartAttempt(context.Background(), task.ID, now)
	require.NoError(t, err)
	require.Equal(t, 1, attempt.AttemptNo)

	nextRunAt := now.Add(time.Minute)
	require.NoError(t, repo.MarkFailed(context.Background(), task.ID, 1, service.ReviewTaskStatusPending, &nextRunAt, "temporary", nil))
	failedOnce, err := repo.FindByID(context.Background(), task.ID)
	require.NoError(t, err)
	require.Equal(t, 1, failedOnce.Attempts)
	require.Equal(t, service.ReviewTaskStatusPending, failedOnce.Status)

	require.NoError(t, repo.MarkSucceeded(context.Background(), task.ID, now.Add(time.Minute*2)))
	succeeded, err := repo.FindByID(context.Background(), task.ID)
	require.NoError(t, err)
	require.Equal(t, service.ReviewTaskStatusSucceeded, succeeded.Status)
}

func openReviewTaskRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.ReviewTask{}, &model.ReviewTaskAttempt{}))
	return db
}
