package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/stretchr/testify/require"
)

func TestStatsRepositoryListReviewStatsEntriesReturnsPushAndMergeRowsInRange(t *testing.T) {
	db := openReviewLogRepositoryTestDB(t)
	repo := NewStatsRepository(db)
	base := time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)

	require.NoError(t, db.Create(&model.PushReviewLog{
		ProjectID:         1,
		ProjectName:       "api",
		Author:            "alice-legacy",
		AuthorIdentity:    "alice",
		AuthorDisplayName: "Alice",
		Score:             90,
		Additions:         12,
		Deletions:         3,
		CreatedAt:         base,
		UpdatedAt:         base,
	}).Error)
	require.NoError(t, db.Create(&model.MergeRequestReviewLog{
		ProjectID:         2,
		ProjectName:       "web",
		AuthorIdentity:    "bob",
		AuthorDisplayName: "Bob",
		Score:             70,
		Additions:         4,
		Deletions:         1,
		CreatedAt:         base.Add(time.Hour),
		UpdatedAt:         base.Add(time.Hour),
	}).Error)
	require.NoError(t, db.Create(&model.PushReviewLog{
		ProjectID:      3,
		ProjectName:    "outside",
		AuthorIdentity: "charlie",
		Score:          60,
		CreatedAt:      base.Add(48 * time.Hour),
		UpdatedAt:      base.Add(48 * time.Hour),
	}).Error)

	rows, err := repo.ListReviewStatsEntries(context.Background(), service.StatsEntryQuery{
		StartTime: base.Add(-time.Minute),
		EndTime:   base.Add(2 * time.Hour),
	})

	require.NoError(t, err)
	require.Equal(t, []service.ReviewStatsLogEntry{
		{
			ProjectID:         1,
			ProjectName:       "api",
			Author:            "alice-legacy",
			AuthorIdentity:    "alice",
			AuthorDisplayName: "Alice",
			Score:             90,
			Additions:         12,
			Deletions:         3,
			CreatedAt:         base.UnixMilli(),
		},
		{
			ProjectID:         2,
			ProjectName:       "web",
			AuthorIdentity:    "bob",
			AuthorDisplayName: "Bob",
			Score:             70,
			Additions:         4,
			Deletions:         1,
			CreatedAt:         base.Add(time.Hour).UnixMilli(),
		},
	}, rows)
}

func TestStatsRepositoryListReviewStatsEntriesFiltersProjectName(t *testing.T) {
	db := openReviewLogRepositoryTestDB(t)
	repo := NewStatsRepository(db)
	now := time.Now().UTC()

	require.NoError(t, db.Create(&model.PushReviewLog{
		ProjectID:      1,
		ProjectName:    "api",
		AuthorIdentity: "alice",
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error)
	require.NoError(t, db.Create(&model.MergeRequestReviewLog{
		ProjectID:      2,
		ProjectName:    "web",
		AuthorIdentity: "bob",
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error)

	rows, err := repo.ListReviewStatsEntries(context.Background(), service.StatsEntryQuery{
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Minute),
		Project:   "web",
	})

	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "web", rows[0].ProjectName)
}
