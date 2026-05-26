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

func TestSysLogRepositoryGetAndSearch(t *testing.T) {
	db := openSysLogRepositoryTestDB(t)
	repo := NewSysLogRepository(db)
	now := time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)

	require.NoError(t, db.Create(&model.SysLog{
		Level:      "INFO",
		Module:     "REVIEW",
		Action:     "send-webhook",
		Message:    "notification sent",
		Detail:     `{"projectId":1}`,
		ErrorStack: "",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error)
	require.NoError(t, db.Create(&model.SysLog{
		Level:     "ERROR",
		Module:    "SYSTEM",
		Action:    "boot",
		Message:   "startup failed",
		CreatedAt: now.Add(time.Hour),
		UpdatedAt: now.Add(time.Hour),
	}).Error)

	got, err := repo.FindSysLogByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "send-webhook", got.Action)
	require.Equal(t, now.UnixMilli(), got.CreatedAt)

	page, err := repo.SearchSysLogs(context.Background(), service.SysLogSearchQuery{
		Level:   "INFO",
		Module:  "REVIEW",
		Action:  "web",
		Message: "sent",
		Page:    1,
		Size:    20,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, `{"projectId":1}`, page.Items[0].Detail)

	_, err = repo.FindSysLogByID(context.Background(), 99)
	require.ErrorIs(t, err, service.ErrSysLogNotFound)
}

func openSysLogRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.SysLog{}))
	return db
}
