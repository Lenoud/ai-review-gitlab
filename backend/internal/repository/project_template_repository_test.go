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

func TestProjectTemplateRepositoryCreatesUpdatesAndFindsTemplate(t *testing.T) {
	db := openProjectTemplateRepositoryTestDB(t)
	repo := NewProjectTemplateRepository(db)

	created, err := repo.CreateProjectTemplate(context.Background(), service.ProjectTemplateInput{
		Name:                 "Go service",
		Description:          "backend",
		Extensions:           []string{".go", ".mod"},
		ReviewPromptTemplate: "review",
	})

	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, "Go service", created.Name)
	require.Equal(t, []string{".go", ".mod"}, created.Extensions)

	updated, err := repo.UpdateProjectTemplate(context.Background(), created.ID, service.ProjectTemplateInput{
		Name:                 "Vue app",
		Description:          "frontend",
		Extensions:           []string{".vue", ".ts"},
		ReviewPromptTemplate: "review ui",
	})

	require.NoError(t, err)
	require.Equal(t, "Vue app", updated.Name)
	require.Equal(t, "frontend", updated.Description)
	require.Equal(t, []string{".vue", ".ts"}, updated.Extensions)

	found, err := repo.FindProjectTemplateByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, "Vue app", found.Name)
}

func TestProjectTemplateRepositoryFindMapsNotFound(t *testing.T) {
	db := openProjectTemplateRepositoryTestDB(t)
	repo := NewProjectTemplateRepository(db)

	_, err := repo.FindProjectTemplateByID(context.Background(), 404)

	require.ErrorIs(t, err, service.ErrProjectTemplateNotFound)
}

func TestProjectTemplateRepositoryDeletesTemplates(t *testing.T) {
	db := openProjectTemplateRepositoryTestDB(t)
	record := model.ProjectTemplate{Name: "Go service"}
	require.NoError(t, db.Create(&record).Error)
	rule := model.ProjectTemplateReviewRule{TemplateID: record.ID, Name: "rule", GlobPatterns: []byte(`["*.go"]`), Content: "content", Enabled: true}
	require.NoError(t, db.Create(&rule).Error)
	repo := NewProjectTemplateRepository(db)

	err := repo.DeleteProjectTemplates(context.Background(), []uint{record.ID})

	require.NoError(t, err)
	_, err = repo.FindProjectTemplateByID(context.Background(), record.ID)
	require.ErrorIs(t, err, service.ErrProjectTemplateNotFound)

	var count int64
	require.NoError(t, db.Model(&model.ProjectTemplateReviewRule{}).Where("template_id = ?", record.ID).Count(&count).Error)
	require.Zero(t, count)
}

func TestProjectTemplateRepositoryListsTemplates(t *testing.T) {
	db := openProjectTemplateRepositoryTestDB(t)
	require.NoError(t, db.Exec(`
		INSERT INTO project_template (name, description, extensions, review_prompt_template)
		VALUES
			('Go service', 'backend API', '[".go"]', 'go review'),
			('Vue app', 'frontend', '[".vue"]', 'vue review'),
			('Full stack', 'Go and Vue', '[".go",".vue"]', 'general review')
	`).Error)
	repo := NewProjectTemplateRepository(db)

	items, err := repo.ListProjectTemplates(context.Background(), service.ProjectTemplateListQuery{
		Keyword: "Go",
	})

	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "Full stack", items[0].Name)
	require.Equal(t, "Go service", items[1].Name)
}

func TestProjectTemplateRepositoryCountsProjectsUsingTemplates(t *testing.T) {
	db := openProjectTemplateRepositoryTestDB(t)
	template := model.ProjectTemplate{Name: "Go service"}
	require.NoError(t, db.Create(&template).Error)
	require.NoError(t, db.Create(&model.Project{
		Name:       "api",
		WebURL:     "https://gitlab.example.com/group/api",
		Platform:   "gitlab",
		TemplateID: template.ID,
	}).Error)
	repo := NewProjectTemplateRepository(db)

	count, err := repo.CountProjectsUsingTemplates(context.Background(), []uint{template.ID})

	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func openProjectTemplateRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.ProjectTemplate{}, &model.ProjectTemplateReviewRule{}, &model.Project{}))
	return db
}
