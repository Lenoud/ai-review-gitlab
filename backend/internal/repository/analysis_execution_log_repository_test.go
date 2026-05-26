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

func TestReviewLogRepositoryFindsAnalysisExecutionByID(t *testing.T) {
	db := openAnalysisExecutionLogRepositoryTestDB(t)
	startedAt := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(2 * time.Minute)
	record := model.ProjectAnalysisPlanExecutionLog{
		ProjectID:           7,
		PlanID:              3,
		Status:              "succeeded",
		ResultContent:       "整体趋势稳定",
		ShareToken:          "analysis-token",
		ShareTokenExpiresAt: completedAt.Add(7 * 24 * time.Hour).UnixMilli(),
		StartedAt:           startedAt,
		CompletedAt:         completedAt,
		DurationMs:          120000,
	}
	require.NoError(t, db.Create(&record).Error)
	repo := NewReviewLogRepository(db)

	got, err := repo.FindAnalysisExecutionByID(context.Background(), record.ID)

	require.NoError(t, err)
	require.Equal(t, record.ID, got.ID)
	require.Equal(t, uint(7), got.ProjectID)
	require.Equal(t, uint(3), got.PlanID)
	require.Equal(t, "succeeded", got.Status)
	require.Equal(t, "整体趋势稳定", got.ResultContent)
	require.Equal(t, "analysis-token", got.ShareToken)
	require.Equal(t, int64(120000), got.DurationMs)
	require.Equal(t, completedAt, got.CompletedAt)
}

func TestReviewLogRepositoryFindAnalysisExecutionByIDMapsNotFound(t *testing.T) {
	db := openAnalysisExecutionLogRepositoryTestDB(t)
	repo := NewReviewLogRepository(db)

	_, err := repo.FindAnalysisExecutionByID(context.Background(), 404)

	require.ErrorIs(t, err, service.ErrReviewLogNotFound)
}

func openAnalysisExecutionLogRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.ProjectAnalysisPlanExecutionLog{}))
	return db
}
