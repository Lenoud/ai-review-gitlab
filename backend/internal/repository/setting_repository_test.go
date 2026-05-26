package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSettingRepositoryReturnsMissingValue(t *testing.T) {
	db := openSettingRepositoryTestDB(t)
	repo := NewSettingRepository(db)

	value, found, err := repo.GetSettingValue(context.Background(), "SYSTEM")

	require.NoError(t, err)
	require.False(t, found)
	require.Empty(t, value)
}

func TestSettingRepositoryCreatesAndUpdatesValue(t *testing.T) {
	db := openSettingRepositoryTestDB(t)
	repo := NewSettingRepository(db)

	require.NoError(t, repo.SetSettingValue(context.Background(), "SYSTEM", `{"siteName":"AI Code Review"}`))
	value, found, err := repo.GetSettingValue(context.Background(), "SYSTEM")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, `{"siteName":"AI Code Review"}`, value)

	require.NoError(t, repo.SetSettingValue(context.Background(), "SYSTEM", `{"siteName":"AI Review","baseUrl":"https://cr.example.com"}`))
	value, found, err = repo.GetSettingValue(context.Background(), "SYSTEM")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, `{"siteName":"AI Review","baseUrl":"https://cr.example.com"}`, value)

	var count int64
	require.NoError(t, db.Model(&model.Setting{}).Where("`key` = ?", "SYSTEM").Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestSettingRepositorySetRetriesDuplicateCreateAsUpdate(t *testing.T) {
	db := openSettingRepositoryTestDB(t)
	repo := NewSettingRepository(db)

	require.NoError(t, db.Create(&model.Setting{
		Key:   "SYSTEM",
		Value: `{"siteName":"first"}`,
	}).Error)

	err := repo.setSettingValueAfterMissing(context.Background(), "SYSTEM", `{"siteName":"second"}`)

	require.NoError(t, err)
	value, found, err := repo.GetSettingValue(context.Background(), "SYSTEM")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, `{"siteName":"second"}`, value)
}

func openSettingRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Setting{}))
	return db
}
