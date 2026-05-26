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

func TestMemberIMMappingRepositoryCreatesUpdatesAndFindsMapping(t *testing.T) {
	db := openMemberIMMappingRepositoryTestDB(t)
	repo := NewMemberIMMappingRepository(db)

	created, err := repo.CreateMemberIMMapping(context.Background(), service.MemberIMMappingInput{
		GitUsername: "alice",
		Platform:    service.IMRobotPlatformDingTalk,
		IMUserID:    "ding-user",
		DisplayName: "Alice",
	})

	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, "alice", created.GitUsername)

	updated, err := repo.UpdateMemberIMMapping(context.Background(), created.ID, service.MemberIMMappingInput{
		GitUsername: "alice",
		Platform:    service.IMRobotPlatformFeishu,
		IMUserID:    "ou_123",
		DisplayName: "Alice Chen",
	})

	require.NoError(t, err)
	require.Equal(t, service.IMRobotPlatformFeishu, updated.Platform)
	require.Equal(t, "ou_123", updated.IMUserID)
	require.Equal(t, "Alice Chen", updated.DisplayName)

	found, err := repo.FindMemberIMMappingByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, "alice", found.GitUsername)
	require.Equal(t, service.IMRobotPlatformFeishu, found.Platform)
}

func TestMemberIMMappingRepositoryIdempotentUpdateReturnsExistingMapping(t *testing.T) {
	db := openMemberIMMappingRepositoryTestDB(t)
	record := insertMemberIMMappingRecord(t, db, model.MemberIMMapping{
		GitUsername: "alice",
		Platform:    service.IMRobotPlatformDingTalk,
		IMUserID:    "ding-1",
		DisplayName: "Alice",
	})
	repo := NewMemberIMMappingRepository(db)

	updated, err := repo.UpdateMemberIMMapping(context.Background(), record.ID, service.MemberIMMappingInput{
		GitUsername: "alice",
		Platform:    service.IMRobotPlatformDingTalk,
		IMUserID:    "ding-1",
		DisplayName: "Alice",
	})

	require.NoError(t, err)
	require.Equal(t, record.ID, updated.ID)
	require.Equal(t, "alice", updated.GitUsername)
}

func TestMemberIMMappingRepositoryFindMapsNotFound(t *testing.T) {
	db := openMemberIMMappingRepositoryTestDB(t)
	repo := NewMemberIMMappingRepository(db)

	_, err := repo.FindMemberIMMappingByID(context.Background(), 404)

	require.ErrorIs(t, err, service.ErrMemberIMMappingNotFound)
}

func TestMemberIMMappingRepositoryUpdateMapsNotFound(t *testing.T) {
	db := openMemberIMMappingRepositoryTestDB(t)
	repo := NewMemberIMMappingRepository(db)

	_, err := repo.UpdateMemberIMMapping(context.Background(), 404, service.MemberIMMappingInput{
		GitUsername: "alice",
		Platform:    service.IMRobotPlatformDingTalk,
		IMUserID:    "ding-user",
	})

	require.ErrorIs(t, err, service.ErrMemberIMMappingNotFound)
}

func TestMemberIMMappingRepositorySearchesMappings(t *testing.T) {
	db := openMemberIMMappingRepositoryTestDB(t)
	insertMemberIMMappingRecord(t, db, model.MemberIMMapping{GitUsername: "alice", Platform: service.IMRobotPlatformDingTalk, IMUserID: "ding-1", DisplayName: "Alice"})
	insertMemberIMMappingRecord(t, db, model.MemberIMMapping{GitUsername: "bob", Platform: service.IMRobotPlatformDingTalk, IMUserID: "ding-2", DisplayName: "Bob"})
	insertMemberIMMappingRecord(t, db, model.MemberIMMapping{GitUsername: "alice", Platform: service.IMRobotPlatformFeishu, IMUserID: "ou_1", DisplayName: "Alice"})
	repo := NewMemberIMMappingRepository(db)

	page, err := repo.SearchMemberIMMappings(context.Background(), service.MemberIMMappingSearchQuery{
		Keyword:  "alice",
		Platform: service.IMRobotPlatformDingTalk,
		Page:     1,
		Size:     20,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "ding-1", page.Items[0].IMUserID)
}

func TestMemberIMMappingRepositoryExistsExcludesID(t *testing.T) {
	db := openMemberIMMappingRepositoryTestDB(t)
	record := insertMemberIMMappingRecord(t, db, model.MemberIMMapping{GitUsername: "alice", Platform: service.IMRobotPlatformDingTalk, IMUserID: "ding-1"})
	repo := NewMemberIMMappingRepository(db)

	exists, err := repo.ExistsMemberIMMapping(context.Background(), "alice", service.IMRobotPlatformDingTalk, 0)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsMemberIMMapping(context.Background(), "alice", service.IMRobotPlatformDingTalk, record.ID)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestMemberIMMappingRepositoryDeletesMappings(t *testing.T) {
	db := openMemberIMMappingRepositoryTestDB(t)
	record := insertMemberIMMappingRecord(t, db, model.MemberIMMapping{GitUsername: "alice", Platform: service.IMRobotPlatformDingTalk, IMUserID: "ding-1"})
	repo := NewMemberIMMappingRepository(db)

	err := repo.DeleteMemberIMMappings(context.Background(), []uint{record.ID})

	require.NoError(t, err)
	_, err = repo.FindMemberIMMappingByID(context.Background(), record.ID)
	require.ErrorIs(t, err, service.ErrMemberIMMappingNotFound)
}

func openMemberIMMappingRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MemberIMMapping{}))
	return db
}

func insertMemberIMMappingRecord(t *testing.T, db *gorm.DB, record model.MemberIMMapping) model.MemberIMMapping {
	t.Helper()

	require.NoError(t, db.Create(&record).Error)
	return record
}
