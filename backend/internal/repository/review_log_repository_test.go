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

func TestReviewLogRepositoryPushCRUDAndSearch(t *testing.T) {
	db := openReviewLogRepositoryTestDB(t)
	repo := NewReviewLogRepository(db)

	log, err := repo.CreatePush(context.Background(), service.PushReviewLogInput{
		ProjectID:         7,
		ProjectName:       "ai-review",
		Author:            "alice",
		AuthorIdentity:    "alice",
		AuthorDisplayName: "Alice",
		Branch:            "main",
		CommitMessages:    "fix auth (by Alice);",
		Commits: []service.ReviewCommit{{
			Author:    "Alice",
			Message:   "fix auth",
			URL:       "https://gitlab.example.com/group/repo/-/commit/abc",
			Timestamp: "2026-05-25T10:00:00Z",
		}},
		Score:         86,
		Additions:     12,
		Deletions:     3,
		LastCommitURL: "https://gitlab.example.com/group/repo/-/commit/abc",
		ReviewResult:  "总分：86分\n整体可合并",
	})

	require.NoError(t, err)
	require.NotZero(t, log.ID)
	require.Equal(t, int64(86), int64(log.Score))
	require.Equal(t, "alice（Alice）", log.AuthorDisplayText)
	require.Len(t, log.Commits, 1)
	require.NotZero(t, log.CreatedAt)

	got, err := repo.FindPushByID(context.Background(), log.ID)
	require.NoError(t, err)
	require.Equal(t, "总分：86分\n整体可合并", got.ReviewResult)

	minScore := 80
	page, err := repo.SearchPush(context.Background(), service.ReviewLogSearchQuery{
		ProjectID:      7,
		Authors:        []string{"alice"},
		ProjectNames:   []string{"ai-review"},
		Branch:         "mai",
		CommitMessages: "auth",
		MinScore:       &minScore,
		Page:           1,
		Size:           10,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, log.ID, page.Items[0].ID)
}

func TestReviewLogRepositoryMergeRequestCRUDAndSearch(t *testing.T) {
	db := openReviewLogRepositoryTestDB(t)
	repo := NewReviewLogRepository(db)

	log, err := repo.CreateMergeRequest(context.Background(), service.MergeRequestReviewLogInput{
		ProjectID:         9,
		ProjectName:       "ai-review",
		Author:            "bob",
		AuthorIdentity:    "bob",
		AuthorDisplayName: "Bob",
		SourceBranch:      "feature/login",
		TargetBranch:      "main",
		CommitMessages:    "add login",
		Score:             72,
		Additions:         22,
		Deletions:         5,
		LastCommitID:      "def456",
		URL:               "https://gitlab.example.com/group/repo/-/merge_requests/5",
		ReviewResult:      "总分:72分",
	})

	require.NoError(t, err)
	require.NotZero(t, log.ID)

	got, err := repo.FindMergeRequestByID(context.Background(), log.ID)
	require.NoError(t, err)
	require.Equal(t, "feature/login", got.SourceBranch)

	maxScore := 80
	start := time.UnixMilli(log.CreatedAt - 1000)
	end := time.UnixMilli(log.CreatedAt + 1000)
	page, err := repo.SearchMergeRequest(context.Background(), service.ReviewLogSearchQuery{
		ProjectID:      9,
		Authors:        []string{"bob"},
		Branch:         "feature",
		CommitMessages: "login",
		MaxScore:       &maxScore,
		StartTime:      &start,
		EndTime:        &end,
		Page:           1,
		Size:           10,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "def456", page.Items[0].LastCommitID)
}

func openReviewLogRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.PushReviewLog{}, &model.MergeRequestReviewLog{}))
	return db
}
