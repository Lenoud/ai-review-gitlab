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
	require.Equal(t, "", got.ResultActions)
	require.Equal(t, "analysis-token", got.ShareToken)
	require.Equal(t, int64(120000), got.DurationMs)
	require.Equal(t, completedAt, got.CompletedAt)
}

func TestReviewLogRepositoryCreatesAnalysisExecution(t *testing.T) {
	db := openAnalysisExecutionLogRepositoryTestDB(t)
	repo := NewReviewLogRepository(db)
	startedAt := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(3 * time.Second)

	got, err := repo.CreateAnalysisExecution(context.Background(), service.AnalysisExecutionLogInput{
		ProjectID:     7,
		PlanID:        3,
		Status:        service.AnalysisExecutionStatusFailed,
		StartedAt:     startedAt,
		CompletedAt:   completedAt,
		DurationMs:    3000,
		ResultContent: "partial",
		ResultActions: `[{"type":"open"}]`,
		ErrorMessage:  "llm unavailable",
		ErrorStack:    "stack trace",
	})

	require.NoError(t, err)
	require.NotZero(t, got.ID)
	require.Equal(t, uint(7), got.ProjectID)
	require.Equal(t, uint(3), got.PlanID)
	require.Equal(t, service.AnalysisExecutionStatusFailed, got.Status)
	require.Equal(t, "partial", got.ResultContent)
	require.Equal(t, `[{"type":"open"}]`, got.ResultActions)
	require.Equal(t, "llm unavailable", got.ErrorMessage)
	require.Equal(t, "stack trace", got.ErrorStack)
	require.Equal(t, startedAt, got.StartedAt)
	require.Equal(t, completedAt, got.CompletedAt)
	require.Equal(t, int64(3000), got.DurationMs)
}

func TestReviewLogRepositoryFindAnalysisExecutionByIDMapsNotFound(t *testing.T) {
	db := openAnalysisExecutionLogRepositoryTestDB(t)
	repo := NewReviewLogRepository(db)

	_, err := repo.FindAnalysisExecutionByID(context.Background(), 404)

	require.ErrorIs(t, err, service.ErrReviewLogNotFound)
}

func TestReviewLogRepositorySearchesAnalysisExecutionLogs(t *testing.T) {
	db := openAnalysisExecutionLogRepositoryTestDB(t)
	now := time.Date(2026, 5, 26, 11, 0, 0, 0, time.UTC)
	records := []model.ProjectAnalysisPlanExecutionLog{
		{ProjectID: 7, PlanID: 3, Status: "succeeded", ResultContent: "newest", StartedAt: now.Add(-time.Hour), CompletedAt: now},
		{ProjectID: 7, PlanID: 3, Status: "failed", ResultContent: "wrong status", StartedAt: now.Add(-2 * time.Hour), CompletedAt: now.Add(-time.Hour)},
		{ProjectID: 8, PlanID: 3, Status: "succeeded", ResultContent: "wrong project", StartedAt: now.Add(-3 * time.Hour), CompletedAt: now.Add(-2 * time.Hour)},
		{ProjectID: 7, PlanID: 4, Status: "succeeded", ResultContent: "wrong plan", StartedAt: now.Add(-4 * time.Hour), CompletedAt: now.Add(-3 * time.Hour)},
	}
	require.NoError(t, db.Create(&records).Error)
	repo := NewReviewLogRepository(db)

	page, err := repo.SearchAnalysisExecution(context.Background(), service.AnalysisExecutionLogSearchQuery{
		ProjectID: 7,
		PlanID:    3,
		Status:    "succeeded",
		Page:      1,
		Size:      10,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "newest", page.Items[0].ResultContent)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 10, page.Size)
}

func TestReviewLogRepositoryUpdatesAnalysisExecutionShareToken(t *testing.T) {
	db := openAnalysisExecutionLogRepositoryTestDB(t)
	record := model.ProjectAnalysisPlanExecutionLog{
		ProjectID:     7,
		PlanID:        3,
		Status:        "succeeded",
		ResultContent: "analysis",
	}
	require.NoError(t, db.Create(&record).Error)
	repo := NewReviewLogRepository(db)

	got, err := repo.UpdateAnalysisExecutionShareToken(context.Background(), record.ID, "share-token", 1770000000000)

	require.NoError(t, err)
	require.Equal(t, "share-token", got.ShareToken)
	require.Equal(t, int64(1770000000000), got.ShareTokenExpiresAt)

	var stored model.ProjectAnalysisPlanExecutionLog
	require.NoError(t, db.First(&stored, record.ID).Error)
	require.Equal(t, "share-token", stored.ShareToken)
}

func openAnalysisExecutionLogRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.ProjectAnalysisPlanExecutionLog{}))
	return db
}
