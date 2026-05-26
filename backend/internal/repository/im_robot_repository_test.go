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

func TestIMRobotRepositoryCreatesUpdatesAndFindsRobot(t *testing.T) {
	db := openIMRobotRepositoryTestDB(t)
	repo := NewIMRobotRepository(db)

	created, err := repo.CreateIMRobot(context.Background(), service.IMRobotInput{
		Platform:   service.IMRobotPlatformDingTalk,
		Name:       "alerts",
		WebhookURL: "https://example.com/dingtalk",
		Secret:     "secret",
		Enabled:    true,
	})

	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, "alerts", created.Name)
	require.True(t, created.Enabled)

	updated, err := repo.UpdateIMRobot(context.Background(), created.ID, service.IMRobotInput{
		Platform:   service.IMRobotPlatformFeishu,
		Name:       "feishu alerts",
		WebhookURL: "https://example.com/feishu",
		Secret:     "new-secret",
		Enabled:    false,
	})

	require.NoError(t, err)
	require.Equal(t, service.IMRobotPlatformFeishu, updated.Platform)
	require.Equal(t, "feishu alerts", updated.Name)
	require.False(t, updated.Enabled)

	found, err := repo.FindIMRobotByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.False(t, found.Enabled)
	require.Equal(t, "new-secret", found.Secret)
}

func TestIMRobotRepositoryFindMapsNotFound(t *testing.T) {
	db := openIMRobotRepositoryTestDB(t)
	repo := NewIMRobotRepository(db)

	_, err := repo.FindIMRobotByID(context.Background(), 404)

	require.ErrorIs(t, err, service.ErrIMRobotNotFound)
}

func TestIMRobotRepositorySearchesRobots(t *testing.T) {
	db := openIMRobotRepositoryTestDB(t)
	insertIMRobotRecord(t, db, model.IMRobot{Platform: service.IMRobotPlatformDingTalk, Name: "build alerts", WebhookURL: "https://example.com/1", Enabled: true})
	insertIMRobotRecord(t, db, model.IMRobot{Platform: service.IMRobotPlatformDingTalk, Name: "deploy alerts", WebhookURL: "https://example.com/2", Enabled: false})
	insertIMRobotRecord(t, db, model.IMRobot{Platform: service.IMRobotPlatformFeishu, Name: "team chat", WebhookURL: "https://example.com/3", Enabled: true})
	repo := NewIMRobotRepository(db)
	enabled := true

	page, err := repo.SearchIMRobots(context.Background(), service.IMRobotSearchQuery{
		Keyword:  "alerts",
		Platform: service.IMRobotPlatformDingTalk,
		Enabled:  &enabled,
		Page:     1,
		Size:     20,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "build alerts", page.Items[0].Name)
}

func TestIMRobotRepositoryListsEnabledRobots(t *testing.T) {
	db := openIMRobotRepositoryTestDB(t)
	insertIMRobotRecord(t, db, model.IMRobot{Platform: service.IMRobotPlatformDingTalk, Name: "disabled", WebhookURL: "https://example.com/1", Enabled: false})
	insertIMRobotRecord(t, db, model.IMRobot{Platform: service.IMRobotPlatformFeishu, Name: "enabled", WebhookURL: "https://example.com/2", Enabled: true})
	repo := NewIMRobotRepository(db)

	items, err := repo.ListEnabledIMRobots(context.Background())

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "enabled", items[0].Name)
}

func TestIMRobotRepositoryDeletesRobots(t *testing.T) {
	db := openIMRobotRepositoryTestDB(t)
	record := model.IMRobot{Platform: service.IMRobotPlatformDingTalk, Name: "alerts", WebhookURL: "https://example.com/1", Enabled: true}
	require.NoError(t, db.Create(&record).Error)
	repo := NewIMRobotRepository(db)

	err := repo.DeleteIMRobots(context.Background(), []uint{record.ID})

	require.NoError(t, err)
	_, err = repo.FindIMRobotByID(context.Background(), record.ID)
	require.ErrorIs(t, err, service.ErrIMRobotNotFound)
}

func TestIMRobotRepositoryCountsRobotReferences(t *testing.T) {
	db := openIMRobotRepositoryTestDB(t)
	robot := model.IMRobot{Platform: service.IMRobotPlatformDingTalk, Name: "alerts", WebhookURL: "https://example.com/1", Enabled: true}
	require.NoError(t, db.Create(&robot).Error)
	require.NoError(t, db.Create(&model.Project{
		Name:      "api",
		WebURL:    "https://gitlab.example.com/group/api",
		Platform:  "gitlab",
		IMRobotID: robot.ID,
	}).Error)
	require.NoError(t, db.Create(&model.ProjectAnalysisPlan{
		ProjectID:      robot.ID,
		Name:           "weekly",
		CronExpression: "0 0 * * *",
		IMRobotID:      robot.ID,
	}).Error)
	repo := NewIMRobotRepository(db)

	count, err := repo.CountIMRobotReferences(context.Background(), []uint{robot.ID})

	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func openIMRobotRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.IMRobot{}, &model.Project{}, &model.ProjectAnalysisPlan{}))
	return db
}

func insertIMRobotRecord(t *testing.T, db *gorm.DB, record model.IMRobot) {
	t.Helper()

	enabled := record.Enabled
	require.NoError(t, db.Create(&record).Error)
	if !enabled {
		require.NoError(t, db.Model(&model.IMRobot{}).Where("id = ?", record.ID).Update("enabled", false).Error)
	}
}
