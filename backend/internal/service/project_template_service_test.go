package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectTemplateServiceCreateNormalizesInput(t *testing.T) {
	repo := &fakeProjectTemplateRepository{
		create: func(ctx context.Context, input ProjectTemplateInput) (*ProjectTemplate, error) {
			require.Equal(t, "Go service", input.Name)
			require.Equal(t, "backend defaults", input.Description)
			require.Equal(t, []string{".go", ".mod"}, input.Extensions)
			require.Equal(t, "review {{diff}}", input.ReviewPromptTemplate)
			return &ProjectTemplate{ID: 1, Name: input.Name, Extensions: input.Extensions}, nil
		},
	}
	svc := NewProjectTemplateService(repo)

	template, err := svc.Create(context.Background(), ProjectTemplateInput{
		Name:                 " Go service ",
		Description:          " backend defaults ",
		Extensions:           []string{" .go ", "", ".mod", ".go"},
		ReviewPromptTemplate: " review {{diff}} ",
	})

	require.NoError(t, err)
	require.Equal(t, uint(1), template.ID)
	require.Equal(t, []string{".go", ".mod"}, template.Extensions)
}

func TestProjectTemplateServiceRejectsInvalidCreate(t *testing.T) {
	svc := NewProjectTemplateService(&fakeProjectTemplateRepository{})

	_, err := svc.Create(context.Background(), ProjectTemplateInput{Name: " "})

	require.ErrorIs(t, err, ErrInvalidProjectTemplateInput)
}

func TestProjectTemplateServiceUpdateRequiresID(t *testing.T) {
	svc := NewProjectTemplateService(&fakeProjectTemplateRepository{})

	_, err := svc.Update(context.Background(), 0, ProjectTemplateInput{Name: "template"})

	require.ErrorIs(t, err, ErrInvalidProjectTemplateInput)
}

func TestProjectTemplateServiceGetRequiresID(t *testing.T) {
	svc := NewProjectTemplateService(&fakeProjectTemplateRepository{})

	_, err := svc.Get(context.Background(), 0)

	require.ErrorIs(t, err, ErrInvalidProjectTemplateInput)
}

func TestProjectTemplateServiceDeleteCleansIDs(t *testing.T) {
	repo := &fakeProjectTemplateRepository{
		countProjects: func(ctx context.Context, ids []uint) (int64, error) {
			require.Equal(t, []uint{3, 2}, ids)
			return 0, nil
		},
		delete: func(ctx context.Context, ids []uint) error {
			require.Equal(t, []uint{3, 2}, ids)
			return nil
		},
	}
	svc := NewProjectTemplateService(repo)

	err := svc.Delete(context.Background(), []uint{0, 3, 3, 2})

	require.NoError(t, err)
}

func TestProjectTemplateServiceDeleteRejectsReferencedTemplates(t *testing.T) {
	repo := &fakeProjectTemplateRepository{
		countProjects: func(ctx context.Context, ids []uint) (int64, error) {
			require.Equal(t, []uint{3}, ids)
			return 1, nil
		},
	}
	svc := NewProjectTemplateService(repo)

	err := svc.Delete(context.Background(), []uint{3})

	require.ErrorIs(t, err, ErrProjectTemplateInUse)
}

func TestProjectTemplateServiceDeleteRejectsEmptyIDs(t *testing.T) {
	svc := NewProjectTemplateService(&fakeProjectTemplateRepository{})

	err := svc.Delete(context.Background(), []uint{0, 0})

	require.ErrorIs(t, err, ErrInvalidProjectTemplateInput)
}

func TestProjectTemplateServiceListNormalizesQuery(t *testing.T) {
	repo := &fakeProjectTemplateRepository{
		list: func(ctx context.Context, query ProjectTemplateListQuery) ([]ProjectTemplate, error) {
			require.Equal(t, "Go", query.Keyword)
			return []ProjectTemplate{{ID: 1, Name: "Go service"}}, nil
		},
	}
	svc := NewProjectTemplateService(repo)

	items, err := svc.List(context.Background(), ProjectTemplateListQuery{Keyword: " Go "})

	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestProjectTemplateServicePropagatesNotFound(t *testing.T) {
	svc := NewProjectTemplateService(&fakeProjectTemplateRepository{
		find: func(ctx context.Context, id uint) (*ProjectTemplate, error) {
			return nil, ErrProjectTemplateNotFound
		},
	})

	_, err := svc.Get(context.Background(), 404)

	require.ErrorIs(t, err, ErrProjectTemplateNotFound)
}

type fakeProjectTemplateRepository struct {
	create        func(context.Context, ProjectTemplateInput) (*ProjectTemplate, error)
	update        func(context.Context, uint, ProjectTemplateInput) (*ProjectTemplate, error)
	find          func(context.Context, uint) (*ProjectTemplate, error)
	delete        func(context.Context, []uint) error
	list          func(context.Context, ProjectTemplateListQuery) ([]ProjectTemplate, error)
	countProjects func(context.Context, []uint) (int64, error)
}

func (r *fakeProjectTemplateRepository) CreateProjectTemplate(ctx context.Context, input ProjectTemplateInput) (*ProjectTemplate, error) {
	if r.create == nil {
		return nil, errors.New("unexpected create")
	}
	return r.create(ctx, input)
}

func (r *fakeProjectTemplateRepository) UpdateProjectTemplate(ctx context.Context, id uint, input ProjectTemplateInput) (*ProjectTemplate, error) {
	if r.update == nil {
		return nil, errors.New("unexpected update")
	}
	return r.update(ctx, id, input)
}

func (r *fakeProjectTemplateRepository) FindProjectTemplateByID(ctx context.Context, id uint) (*ProjectTemplate, error) {
	if r.find == nil {
		return nil, errors.New("unexpected find")
	}
	return r.find(ctx, id)
}

func (r *fakeProjectTemplateRepository) DeleteProjectTemplates(ctx context.Context, ids []uint) error {
	if r.delete == nil {
		return errors.New("unexpected delete")
	}
	return r.delete(ctx, ids)
}

func (r *fakeProjectTemplateRepository) ListProjectTemplates(ctx context.Context, query ProjectTemplateListQuery) ([]ProjectTemplate, error) {
	if r.list == nil {
		return nil, errors.New("unexpected list")
	}
	return r.list(ctx, query)
}

func (r *fakeProjectTemplateRepository) CountProjectsUsingTemplates(ctx context.Context, ids []uint) (int64, error) {
	if r.countProjects == nil {
		return 0, errors.New("unexpected count projects")
	}
	return r.countProjects(ctx, ids)
}
