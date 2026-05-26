package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStatsServiceGetStatsAggregatesPushAndMergeLogs(t *testing.T) {
	repo := &fakeStatsRepository{
		rows: []ReviewStatsLogEntry{
			{ProjectID: 1, ProjectName: "api", AuthorIdentity: "alice", AuthorDisplayName: "Alice", Score: 80, Additions: 10, Deletions: 2, CreatedAt: 1000},
			{ProjectID: 1, ProjectName: "api", AuthorIdentity: "alice", AuthorDisplayName: "Alice", Score: 100, Additions: 5, Deletions: 1, CreatedAt: 2000},
			{ProjectID: 2, ProjectName: "web", AuthorIdentity: "bob", AuthorDisplayName: "Bob", Score: 70, Additions: 3, Deletions: 4, CreatedAt: 3000},
		},
	}
	stats, err := NewStatsService(repo).GetStats(context.Background(), StatsRange{StartTime: 1, EndTime: 4000})

	require.NoError(t, err)
	require.Equal(t, 2, stats.ActiveProjects)
	require.Equal(t, 2, stats.Contributors)
	require.Equal(t, int64(3), stats.TotalCommits)
	require.Equal(t, 83.33, stats.AverageScore)
	require.Equal(t, []ProjectStatsItem{
		{ProjectName: "api", CommitCount: 2, AverageScore: 90},
		{ProjectName: "web", CommitCount: 1, AverageScore: 70},
	}, stats.ProjectCommitStats)
	require.Equal(t, []AuthorStatsItem{
		{Author: "alice（Alice）", CommitCount: 2, AverageScore: 90},
		{Author: "bob（Bob）", CommitCount: 1, AverageScore: 70},
	}, stats.AuthorCommitStats)
	require.Equal(t, []CodeChangeStatsItem{
		{Name: "api", Additions: 15, Deletions: 3},
		{Name: "web", Additions: 3, Deletions: 4},
	}, stats.ProjectCodeChangeStats)
	require.Equal(t, []CodeChangeStatsItem{
		{Name: "alice（Alice）", Additions: 15, Deletions: 3},
		{Name: "bob（Bob）", Additions: 3, Deletions: 4},
	}, stats.AuthorCodeChangeStats)
}

func TestStatsServiceRejectsInvalidTimeRange(t *testing.T) {
	_, err := NewStatsService(&fakeStatsRepository{}).GetStats(context.Background(), StatsRange{StartTime: 10, EndTime: 10})

	require.ErrorIs(t, err, ErrInvalidStatsInput)
}

func TestStatsServiceFallsBackToLegacyAuthorWhenIdentityIsMissing(t *testing.T) {
	repo := &fakeStatsRepository{
		rows: []ReviewStatsLogEntry{
			{ProjectID: 1, ProjectName: "api", Author: "legacy-author", Score: 80, CreatedAt: 1000},
		},
	}

	stats, err := NewStatsService(repo).GetStats(context.Background(), StatsRange{StartTime: 1, EndTime: 4000})

	require.NoError(t, err)
	require.Equal(t, 1, stats.Contributors)
	require.Equal(t, []AuthorStatsItem{{Author: "legacy-author", CommitCount: 1, AverageScore: 80}}, stats.AuthorCommitStats)
}

func TestStatsServiceExcludesZeroScoresFromAverages(t *testing.T) {
	repo := &fakeStatsRepository{
		rows: []ReviewStatsLogEntry{
			{ProjectID: 1, ProjectName: "api", AuthorIdentity: "alice", Score: 0, CreatedAt: 1000},
			{ProjectID: 1, ProjectName: "api", AuthorIdentity: "alice", Score: 80, CreatedAt: 2000},
		},
	}
	stats, err := NewStatsService(repo).GetStats(context.Background(), StatsRange{StartTime: 1, EndTime: 4000})

	require.NoError(t, err)
	require.Equal(t, int64(2), stats.TotalCommits)
	require.Equal(t, 80.0, stats.AverageScore)
	require.Equal(t, 80.0, stats.ProjectCommitStats[0].AverageScore)
	require.Equal(t, 80.0, stats.AuthorCommitStats[0].AverageScore)
}

func TestStatsServiceRejectsRangesLongerThanOneYear(t *testing.T) {
	_, err := NewStatsService(&fakeStatsRepository{}).GetStats(context.Background(), StatsRange{
		StartTime: 1000,
		EndTime:   1000 + int64((367*24*time.Hour)/time.Millisecond),
	})

	require.ErrorIs(t, err, ErrInvalidStatsInput)
}

func TestStatsServiceMemberCommitSummaryBuildsDailyBucketsAndPages(t *testing.T) {
	repo := &fakeStatsRepository{
		rows: []ReviewStatsLogEntry{
			{ProjectID: 1, ProjectName: "api", AuthorIdentity: "alice", AuthorDisplayName: "Alice", Score: 80, Additions: 10, Deletions: 2, CreatedAt: 1764547200000},
			{ProjectID: 2, ProjectName: "web", AuthorIdentity: "alice", AuthorDisplayName: "Alice", Score: 100, Additions: 5, Deletions: 1, CreatedAt: 1764633600000},
			{ProjectID: 2, ProjectName: "web", AuthorIdentity: "bob", AuthorDisplayName: "Bob", Score: 70, Additions: 3, Deletions: 4, CreatedAt: 1764633600000},
		},
	}
	page, err := NewStatsService(repo).GetMemberCommitSummary(context.Background(), MemberCommitSummaryQuery{
		StartTime: 1764547200000,
		EndTime:   1764719999000,
		Page:      1,
		Size:      1,
	})

	require.NoError(t, err)
	require.Equal(t, int64(2), page.Total)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 1, page.Size)
	require.Len(t, page.Items, 1)
	require.Equal(t, "alice", page.Items[0].AuthorIdentity)
	require.Equal(t, "alice（Alice）", page.Items[0].AuthorDisplayText)
	require.Equal(t, MemberCommitSummary{
		ProjectCount:         2,
		CommitCount:          2,
		AvgCommitCountPerDay: 0.67,
		AvgScore:             90,
		AdditionCount:        15,
		DeletionCount:        3,
	}, page.Items[0].Summary)
	require.Equal(t, []DailyCommitStats{
		{Date: "2025-12-01", Count: 1},
		{Date: "2025-12-02", Count: 1},
		{Date: "2025-12-03", Count: 0},
	}, page.Items[0].DailyCommits)
}

type fakeStatsRepository struct {
	rows []ReviewStatsLogEntry
}

func (r *fakeStatsRepository) ListReviewStatsEntries(ctx context.Context, query StatsEntryQuery) ([]ReviewStatsLogEntry, error) {
	return r.rows, nil
}
