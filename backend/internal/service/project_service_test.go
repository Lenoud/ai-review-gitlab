package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectServiceCreateProject(t *testing.T) {
	repo := newMemoryProjectRepository()
	svc := NewProjectService(repo)

	project, err := svc.Create(context.Background(), ProjectInput{
		Name:       "AI Review",
		WebURL:     "https://gitlab.example.com/group/ai-review",
		Platform:   ProjectPlatformGitLab,
		Extensions: []string{".go", ".vue"},
	})

	require.NoError(t, err)
	require.NotZero(t, project.ID)
	require.Equal(t, "AI Review", project.Name)
	require.Equal(t, "https://gitlab.example.com/group/ai-review", project.WebURL)
	require.Equal(t, ProjectPlatformGitLab, project.Platform)
	require.True(t, project.AIReviewEnabled)
	require.Equal(t, []string{".go", ".vue"}, project.Extensions)
}

func TestProjectServiceRejectsDuplicateWebURL(t *testing.T) {
	repo := newMemoryProjectRepository(&Project{
		ID:     1,
		Name:   "Existing",
		WebURL: "https://gitlab.example.com/group/ai-review",
	})
	svc := NewProjectService(repo)

	_, err := svc.Create(context.Background(), ProjectInput{
		Name:     "Duplicate",
		WebURL:   "https://gitlab.example.com/group/ai-review",
		Platform: ProjectPlatformGitLab,
	})

	require.ErrorIs(t, err, ErrProjectWebURLExists)
}

func TestProjectServiceUpdateMissingProject(t *testing.T) {
	svc := NewProjectService(newMemoryProjectRepository())

	_, err := svc.Update(context.Background(), 99, ProjectInput{
		Name:     "Missing",
		WebURL:   "https://gitlab.example.com/group/missing",
		Platform: ProjectPlatformGitLab,
	})

	require.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectServiceDeleteRejectsEmptyIDs(t *testing.T) {
	svc := NewProjectService(newMemoryProjectRepository())

	err := svc.Delete(context.Background(), nil)

	require.ErrorIs(t, err, ErrInvalidProjectInput)
}

func TestProjectServiceSearchAndURLExists(t *testing.T) {
	repo := newMemoryProjectRepository(
		&Project{ID: 1, Name: "AI Review", WebURL: "https://gitlab.example.com/group/ai-review", Platform: ProjectPlatformGitLab},
		&Project{ID: 2, Name: "Other", WebURL: "https://gitlab.example.com/group/other", Platform: ProjectPlatformGitLab},
	)
	svc := NewProjectService(repo)

	result, err := svc.Search(context.Background(), ProjectSearchQuery{
		Keyword: "review",
		Page:    1,
		Size:    10,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), result.Total)
	require.Len(t, result.Items, 1)
	require.Equal(t, "AI Review", result.Items[0].Name)

	exists, err := svc.WebURLExists(context.Background(), "https://gitlab.example.com/group/ai-review", 0)
	require.NoError(t, err)
	require.True(t, exists)
}

type memoryProjectRepository struct {
	projects map[uint]*Project
	nextID   uint
}

func newMemoryProjectRepository(projects ...*Project) *memoryProjectRepository {
	repo := &memoryProjectRepository{projects: map[uint]*Project{}, nextID: 1}
	for _, project := range projects {
		copy := *project
		repo.projects[copy.ID] = &copy
		if copy.ID >= repo.nextID {
			repo.nextID = copy.ID + 1
		}
	}
	return repo
}

func (r *memoryProjectRepository) Create(ctx context.Context, input ProjectInput) (*Project, error) {
	project := projectFromInput(input)
	project.ID = r.nextID
	r.nextID++
	r.projects[project.ID] = project
	return cloneProject(project), nil
}

func (r *memoryProjectRepository) BatchCreate(ctx context.Context, inputs []ProjectInput) ([]Project, error) {
	projects := make([]Project, 0, len(inputs))
	for _, input := range inputs {
		project, err := r.Create(ctx, input)
		if err != nil {
			return nil, err
		}
		projects = append(projects, *project)
	}
	return projects, nil
}

func (r *memoryProjectRepository) Update(ctx context.Context, id uint, input ProjectInput) (*Project, error) {
	_, ok := r.projects[id]
	if !ok {
		return nil, ErrProjectNotFound
	}
	updated := projectFromInput(input)
	updated.ID = id
	r.projects[id] = updated
	return cloneProject(updated), nil
}

func (r *memoryProjectRepository) FindByID(ctx context.Context, id uint) (*Project, error) {
	project, ok := r.projects[id]
	if !ok {
		return nil, ErrProjectNotFound
	}
	return cloneProject(project), nil
}

func (r *memoryProjectRepository) Delete(ctx context.Context, ids []uint) error {
	for _, id := range ids {
		delete(r.projects, id)
	}
	return nil
}

func (r *memoryProjectRepository) Search(ctx context.Context, query ProjectSearchQuery) (*ProjectPage, error) {
	items := make([]Project, 0)
	for _, project := range r.projects {
		if query.Keyword == "" || containsFold(project.Name, query.Keyword) || containsFold(project.WebURL, query.Keyword) {
			items = append(items, *cloneProject(project))
		}
	}
	return &ProjectPage{Items: items, Total: int64(len(items)), Page: query.Page, Size: query.Size}, nil
}

func (r *memoryProjectRepository) ExistsByWebURL(ctx context.Context, webURL string, excludeID uint) (bool, error) {
	for _, project := range r.projects {
		if project.WebURL == webURL && project.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

func (r *memoryProjectRepository) ForceError(err error) ProjectRepository {
	return &projectRepositoryWithError{err: err}
}

type projectRepositoryWithError struct {
	err error
}

func (r *projectRepositoryWithError) Create(ctx context.Context, input ProjectInput) (*Project, error) {
	return nil, r.err
}
func (r *projectRepositoryWithError) BatchCreate(ctx context.Context, inputs []ProjectInput) ([]Project, error) {
	return nil, r.err
}
func (r *projectRepositoryWithError) Update(ctx context.Context, id uint, input ProjectInput) (*Project, error) {
	return nil, r.err
}
func (r *projectRepositoryWithError) FindByID(ctx context.Context, id uint) (*Project, error) {
	return nil, r.err
}
func (r *projectRepositoryWithError) Delete(ctx context.Context, ids []uint) error {
	return r.err
}
func (r *projectRepositoryWithError) Search(ctx context.Context, query ProjectSearchQuery) (*ProjectPage, error) {
	return nil, r.err
}
func (r *projectRepositoryWithError) ExistsByWebURL(ctx context.Context, webURL string, excludeID uint) (bool, error) {
	return false, r.err
}

func TestMemoryProjectRepositoryForceErrorSatisfiesInterface(t *testing.T) {
	err := errors.New("forced")
	_, got := newMemoryProjectRepository().ForceError(err).Create(context.Background(), ProjectInput{})
	require.ErrorIs(t, got, err)
}
