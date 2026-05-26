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

func TestProjectAnalysisPlanRepositoryCreatesUpdatesAndFindsPlan(t *testing.T) {
	db := openProjectAnalysisPlanRepositoryTestDB(t)
	repo := NewProjectAnalysisPlanRepository(db)
	enabled := false
	htmlReportEnabled := true

	created, err := repo.CreateProjectAnalysisPlan(context.Background(), service.ProjectAnalysisPlanInput{
		ProjectID:         7,
		Name:              "weekly",
		Prompt:            "summarize",
		CronExpression:    "0 9 * * 1",
		Enabled:           &enabled,
		IMEnabled:         true,
		IMRobotID:         9,
		HTMLReportEnabled: &htmlReportEnabled,
	})

	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.False(t, created.Enabled)
	require.True(t, created.IMEnabled)
	require.True(t, created.HTMLReportEnabled)

	enabled = true
	htmlReportEnabled = false
	updated, err := repo.UpdateProjectAnalysisPlan(context.Background(), created.ID, service.ProjectAnalysisPlanInput{
		ProjectID:         8,
		Name:              "monthly",
		Prompt:            "deeper summary",
		CronExpression:    "0 10 1 * *",
		Enabled:           &enabled,
		HTMLReportEnabled: &htmlReportEnabled,
	})

	require.NoError(t, err)
	require.Equal(t, uint(8), updated.ProjectID)
	require.Equal(t, "monthly", updated.Name)
	require.True(t, updated.Enabled)
	require.False(t, updated.HTMLReportEnabled)

	found, err := repo.FindProjectAnalysisPlanByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, "monthly", found.Name)
}

func TestProjectAnalysisPlanRepositoryFindMapsNotFound(t *testing.T) {
	db := openProjectAnalysisPlanRepositoryTestDB(t)
	repo := NewProjectAnalysisPlanRepository(db)

	_, err := repo.FindProjectAnalysisPlanByID(context.Background(), 404)

	require.ErrorIs(t, err, service.ErrProjectAnalysisPlanNotFound)
}

func TestProjectAnalysisPlanRepositoryDeletesPlans(t *testing.T) {
	db := openProjectAnalysisPlanRepositoryTestDB(t)
	record := model.ProjectAnalysisPlan{ProjectID: 7, Name: "weekly", Enabled: true, HTMLReportEnabled: true}
	require.NoError(t, db.Create(&record).Error)
	repo := NewProjectAnalysisPlanRepository(db)

	err := repo.DeleteProjectAnalysisPlans(context.Background(), []uint{record.ID})

	require.NoError(t, err)
	_, err = repo.FindProjectAnalysisPlanByID(context.Background(), record.ID)
	require.ErrorIs(t, err, service.ErrProjectAnalysisPlanNotFound)
}

func TestProjectAnalysisPlanRepositorySearchesPlans(t *testing.T) {
	db := openProjectAnalysisPlanRepositoryTestDB(t)
	require.NoError(t, db.Exec(`
		INSERT INTO project_analysis_plan (project_id, name, prompt, enabled, html_report_enabled)
		VALUES
			(7, 'weekly security', 'risk', true, true),
			(7, 'disabled weekly', 'risk', false, true),
			(8, 'other project', 'risk', true, true),
			(7, 'daily quality', 'quality', true, true)
	`).Error)
	repo := NewProjectAnalysisPlanRepository(db)
	enabled := true

	page, err := repo.SearchProjectAnalysisPlans(context.Background(), service.ProjectAnalysisPlanSearchQuery{
		ProjectID: 7,
		Keyword:   "weekly",
		Enabled:   &enabled,
		Page:      1,
		Size:      10,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "weekly security", page.Items[0].Name)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 10, page.Size)
}

func openProjectAnalysisPlanRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.ProjectAnalysisPlan{}))
	return db
}
