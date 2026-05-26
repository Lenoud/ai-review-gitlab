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

func TestProjectTemplateReviewRuleRepositoryCreatesUpdatesAndFindsRule(t *testing.T) {
	db := openProjectTemplateReviewRuleRepositoryTestDB(t)
	template := model.ProjectTemplate{Name: "Go service"}
	require.NoError(t, db.Create(&template).Error)
	repo := NewProjectTemplateReviewRuleRepository(db)

	created, err := repo.CreateProjectTemplateReviewRule(context.Background(), service.ProjectTemplateReviewRuleInput{
		TemplateID:   template.ID,
		Name:         "Controller rules",
		Description:  "HTTP handlers",
		GlobPatterns: []string{"*.go", "internal/**/*.go"},
		Content:      "Use context",
		Priority:     10,
		Enabled:      false,
	})

	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, []string{"*.go", "internal/**/*.go"}, created.GlobPatterns)
	require.False(t, created.Enabled)

	updated, err := repo.UpdateProjectTemplateReviewRule(context.Background(), created.ID, service.ProjectTemplateReviewRuleInput{
		TemplateID:   template.ID,
		Name:         "Service rules",
		Description:  "service layer",
		GlobPatterns: []string{"internal/service/*.go"},
		Content:      "Keep logic small",
		Priority:     1,
		Enabled:      true,
	})

	require.NoError(t, err)
	require.Equal(t, "Service rules", updated.Name)
	require.True(t, updated.Enabled)

	disabled, err := repo.UpdateProjectTemplateReviewRule(context.Background(), created.ID, service.ProjectTemplateReviewRuleInput{
		TemplateID:   template.ID,
		Name:         "Disabled rules",
		Description:  "service layer",
		GlobPatterns: []string{"internal/service/*.go"},
		Content:      "Keep logic small",
		Priority:     1,
		Enabled:      false,
	})

	require.NoError(t, err)
	require.Equal(t, "Disabled rules", disabled.Name)
	require.False(t, disabled.Enabled)

	found, err := repo.FindProjectTemplateReviewRuleByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, []string{"internal/service/*.go"}, found.GlobPatterns)
	require.False(t, found.Enabled)
}

func TestProjectTemplateReviewRuleRepositoryListOrdersByPriorityDescCreatedAsc(t *testing.T) {
	db := openProjectTemplateReviewRuleRepositoryTestDB(t)
	template := model.ProjectTemplate{Name: "Go service"}
	require.NoError(t, db.Create(&template).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO project_template_review_rule (template_id, name, glob_patterns, content, priority, enabled, created_at)
		VALUES
			(?, 'low', '["*.md"]', 'low', 1, true, '2026-01-01 00:00:02'),
			(?, 'high-old', '["*.go"]', 'high old', 10, true, '2026-01-01 00:00:01'),
			(?, 'high-new', '["*.ts"]', 'high new', 10, true, '2026-01-01 00:00:03')
	`, template.ID, template.ID, template.ID).Error)
	repo := NewProjectTemplateReviewRuleRepository(db)

	items, err := repo.ListProjectTemplateReviewRulesByTemplateID(context.Background(), template.ID)

	require.NoError(t, err)
	require.Equal(t, []string{"high-old", "high-new", "low"}, []string{items[0].Name, items[1].Name, items[2].Name})
}

func TestProjectTemplateReviewRuleRepositoryMapsNotFound(t *testing.T) {
	db := openProjectTemplateReviewRuleRepositoryTestDB(t)
	repo := NewProjectTemplateReviewRuleRepository(db)

	_, err := repo.FindProjectTemplateReviewRuleByID(context.Background(), 404)

	require.ErrorIs(t, err, service.ErrProjectTemplateReviewRuleNotFound)
}

func TestProjectTemplateReviewRuleRepositoryChecksTemplateExists(t *testing.T) {
	db := openProjectTemplateReviewRuleRepositoryTestDB(t)
	template := model.ProjectTemplate{Name: "Go service"}
	require.NoError(t, db.Create(&template).Error)
	repo := NewProjectTemplateReviewRuleRepository(db)

	exists, err := repo.ProjectTemplateExists(context.Background(), template.ID)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ProjectTemplateExists(context.Background(), 404)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestProjectTemplateReviewRuleRepositoryDeletesRule(t *testing.T) {
	db := openProjectTemplateReviewRuleRepositoryTestDB(t)
	template := model.ProjectTemplate{Name: "Go service"}
	require.NoError(t, db.Create(&template).Error)
	rule := model.ProjectTemplateReviewRule{TemplateID: template.ID, Name: "rule", GlobPatterns: []byte(`["*.go"]`), Content: "content", Enabled: true}
	require.NoError(t, db.Create(&rule).Error)
	repo := NewProjectTemplateReviewRuleRepository(db)

	err := repo.DeleteProjectTemplateReviewRule(context.Background(), rule.ID)

	require.NoError(t, err)
	_, err = repo.FindProjectTemplateReviewRuleByID(context.Background(), rule.ID)
	require.ErrorIs(t, err, service.ErrProjectTemplateReviewRuleNotFound)
}

func openProjectTemplateReviewRuleRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.ProjectTemplate{}, &model.ProjectTemplateReviewRule{}))
	return db
}
