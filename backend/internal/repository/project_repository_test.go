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

func TestProjectRepositoryCRUDAndSearch(t *testing.T) {
	db := openProjectRepositoryTestDB(t)
	repo := NewProjectRepository(db)

	project, err := repo.Create(context.Background(), service.ProjectInput{
		Name:             "AI Review",
		Description:      "review platform",
		WebURL:           "https://gitlab.example.com/group/ai-review",
		Platform:         service.ProjectPlatformGitLab,
		AccessToken:      "token",
		Extensions:       []string{".go", ".vue"},
		ReviewEventTypes: []string{"push", "merge_request"},
	})
	require.NoError(t, err)

	got, err := repo.FindByID(context.Background(), project.ID)
	require.NoError(t, err)
	require.Equal(t, "AI Review", got.Name)
	require.Equal(t, []string{".go", ".vue"}, got.Extensions)

	updated, err := repo.Update(context.Background(), project.ID, service.ProjectInput{
		Name:       "AI Review Updated",
		WebURL:     "https://gitlab.example.com/group/ai-review",
		Platform:   service.ProjectPlatformGitLab,
		Extensions: []string{".go"},
	})
	require.NoError(t, err)
	require.Equal(t, "AI Review Updated", updated.Name)

	page, err := repo.Search(context.Background(), service.ProjectSearchQuery{
		Keyword: "updated",
		Page:    1,
		Size:    10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)

	exists, err := repo.ExistsByWebURL(context.Background(), "https://gitlab.example.com/group/ai-review", 0)
	require.NoError(t, err)
	require.True(t, exists)

	require.NoError(t, repo.Delete(context.Background(), []uint{project.ID}))
	_, err = repo.FindByID(context.Background(), project.ID)
	require.ErrorIs(t, err, service.ErrProjectNotFound)
}

func TestProjectRepositoryBatchCreate(t *testing.T) {
	db := openProjectRepositoryTestDB(t)
	repo := NewProjectRepository(db)

	projects, err := repo.BatchCreate(context.Background(), []service.ProjectInput{
		{Name: "A", WebURL: "https://gitlab.example.com/a", Platform: service.ProjectPlatformGitLab},
		{Name: "B", WebURL: "https://gitlab.example.com/b", Platform: service.ProjectPlatformGitLab},
	})

	require.NoError(t, err)
	require.Len(t, projects, 2)
}

func openProjectRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=private", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Project{}))
	return db
}
