package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectAnalysisPlanServiceCreateNormalizesInput(t *testing.T) {
	repo := &fakeProjectAnalysisPlanRepository{
		create: func(ctx context.Context, input ProjectAnalysisPlanInput) (*ProjectAnalysisPlan, error) {
			require.Equal(t, uint(7), input.ProjectID)
			require.Equal(t, "weekly analysis", input.Name)
			require.Equal(t, "summarize risks", input.Prompt)
			require.Equal(t, "0 9 * * 1", input.CronExpression)
			require.NotNil(t, input.Enabled)
			require.NotNil(t, input.HTMLReportEnabled)
			require.True(t, *input.Enabled)
			require.True(t, *input.HTMLReportEnabled)
			return &ProjectAnalysisPlan{ID: 1, ProjectID: input.ProjectID, Name: input.Name, Enabled: *input.Enabled, HTMLReportEnabled: *input.HTMLReportEnabled}, nil
		},
	}
	svc := NewProjectAnalysisPlanService(repo)

	plan, err := svc.Create(context.Background(), ProjectAnalysisPlanInput{
		ProjectID:      7,
		Name:           " weekly analysis ",
		Prompt:         " summarize risks ",
		CronExpression: " 0 9 * * 1 ",
	})

	require.NoError(t, err)
	require.Equal(t, uint(1), plan.ID)
	require.True(t, plan.Enabled)
}

func TestProjectAnalysisPlanServiceRejectsInvalidCreate(t *testing.T) {
	svc := NewProjectAnalysisPlanService(&fakeProjectAnalysisPlanRepository{})

	_, err := svc.Create(context.Background(), ProjectAnalysisPlanInput{Name: "missing project"})

	require.ErrorIs(t, err, ErrInvalidProjectAnalysisPlanInput)
}

func TestProjectAnalysisPlanServiceUpdateRequiresID(t *testing.T) {
	svc := NewProjectAnalysisPlanService(&fakeProjectAnalysisPlanRepository{})

	_, err := svc.Update(context.Background(), 0, ProjectAnalysisPlanInput{ProjectID: 1, Name: "plan"})

	require.ErrorIs(t, err, ErrInvalidProjectAnalysisPlanInput)
}

func TestProjectAnalysisPlanServiceGetRequiresID(t *testing.T) {
	svc := NewProjectAnalysisPlanService(&fakeProjectAnalysisPlanRepository{})

	_, err := svc.Get(context.Background(), 0)

	require.ErrorIs(t, err, ErrInvalidProjectAnalysisPlanInput)
}

func TestProjectAnalysisPlanServiceDeleteCleansIDs(t *testing.T) {
	repo := &fakeProjectAnalysisPlanRepository{
		delete: func(ctx context.Context, ids []uint) error {
			require.Equal(t, []uint{3, 2}, ids)
			return nil
		},
	}
	svc := NewProjectAnalysisPlanService(repo)

	err := svc.Delete(context.Background(), []uint{0, 3, 3, 2})

	require.NoError(t, err)
}

func TestProjectAnalysisPlanServiceDeleteRejectsEmptyIDs(t *testing.T) {
	svc := NewProjectAnalysisPlanService(&fakeProjectAnalysisPlanRepository{})

	err := svc.Delete(context.Background(), []uint{0, 0})

	require.ErrorIs(t, err, ErrInvalidProjectAnalysisPlanInput)
}

func TestProjectAnalysisPlanServiceSearchNormalizesPagination(t *testing.T) {
	repo := &fakeProjectAnalysisPlanRepository{
		search: func(ctx context.Context, query ProjectAnalysisPlanSearchQuery) (*ProjectAnalysisPlanPage, error) {
			require.Equal(t, "weekly", query.Keyword)
			require.Equal(t, 1, query.Page)
			require.Equal(t, 200, query.Size)
			return &ProjectAnalysisPlanPage{Items: []ProjectAnalysisPlan{}, Total: 0, Page: query.Page, Size: query.Size}, nil
		},
	}
	svc := NewProjectAnalysisPlanService(repo)

	page, err := svc.Search(context.Background(), ProjectAnalysisPlanSearchQuery{Keyword: " weekly ", Page: -1, Size: 999})

	require.NoError(t, err)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 200, page.Size)
}

func TestProjectAnalysisPlanServicePropagatesNotFound(t *testing.T) {
	svc := NewProjectAnalysisPlanService(&fakeProjectAnalysisPlanRepository{
		find: func(ctx context.Context, id uint) (*ProjectAnalysisPlan, error) {
			return nil, ErrProjectAnalysisPlanNotFound
		},
	})

	_, err := svc.Get(context.Background(), 404)

	require.ErrorIs(t, err, ErrProjectAnalysisPlanNotFound)
}

type fakeProjectAnalysisPlanRepository struct {
	create func(context.Context, ProjectAnalysisPlanInput) (*ProjectAnalysisPlan, error)
	update func(context.Context, uint, ProjectAnalysisPlanInput) (*ProjectAnalysisPlan, error)
	find   func(context.Context, uint) (*ProjectAnalysisPlan, error)
	delete func(context.Context, []uint) error
	search func(context.Context, ProjectAnalysisPlanSearchQuery) (*ProjectAnalysisPlanPage, error)
}

func (r *fakeProjectAnalysisPlanRepository) CreateProjectAnalysisPlan(ctx context.Context, input ProjectAnalysisPlanInput) (*ProjectAnalysisPlan, error) {
	if r.create == nil {
		return nil, errors.New("unexpected create")
	}
	return r.create(ctx, input)
}

func (r *fakeProjectAnalysisPlanRepository) UpdateProjectAnalysisPlan(ctx context.Context, id uint, input ProjectAnalysisPlanInput) (*ProjectAnalysisPlan, error) {
	if r.update == nil {
		return nil, errors.New("unexpected update")
	}
	return r.update(ctx, id, input)
}

func (r *fakeProjectAnalysisPlanRepository) FindProjectAnalysisPlanByID(ctx context.Context, id uint) (*ProjectAnalysisPlan, error) {
	if r.find == nil {
		return nil, errors.New("unexpected find")
	}
	return r.find(ctx, id)
}

func (r *fakeProjectAnalysisPlanRepository) DeleteProjectAnalysisPlans(ctx context.Context, ids []uint) error {
	if r.delete == nil {
		return errors.New("unexpected delete")
	}
	return r.delete(ctx, ids)
}

func (r *fakeProjectAnalysisPlanRepository) SearchProjectAnalysisPlans(ctx context.Context, query ProjectAnalysisPlanSearchQuery) (*ProjectAnalysisPlanPage, error) {
	if r.search == nil {
		return nil, errors.New("unexpected search")
	}
	return r.search(ctx, query)
}
