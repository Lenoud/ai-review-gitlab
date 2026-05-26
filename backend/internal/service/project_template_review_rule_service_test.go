package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectTemplateReviewRuleServiceCreateNormalizesInput(t *testing.T) {
	repo := &fakeProjectTemplateReviewRuleRepository{
		templateExists: func(ctx context.Context, templateID uint) (bool, error) {
			require.Equal(t, uint(7), templateID)
			return true, nil
		},
		create: func(ctx context.Context, input ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error) {
			require.Equal(t, uint(7), input.TemplateID)
			require.Equal(t, "Controller rules", input.Name)
			require.Equal(t, "HTTP handlers", input.Description)
			require.Equal(t, []string{"*.go", "internal/**/*.go"}, input.GlobPatterns)
			require.Equal(t, "Use context", input.Content)
			require.Equal(t, 10, input.Priority)
			require.False(t, input.Enabled)
			return &ProjectTemplateReviewRule{ID: 1, TemplateID: input.TemplateID, Name: input.Name, Enabled: input.Enabled}, nil
		},
	}
	svc := NewProjectTemplateReviewRuleService(repo)

	rule, err := svc.Create(context.Background(), ProjectTemplateReviewRuleInput{
		TemplateID:   7,
		Name:         " Controller rules ",
		Description:  " HTTP handlers ",
		GlobPatterns: []string{" *.go ", "", "internal/**/*.go", "*.go"},
		Content:      " Use context ",
		Priority:     10,
		Enabled:      false,
		EnabledIsSet: true,
	})

	require.NoError(t, err)
	require.Equal(t, uint(1), rule.ID)
	require.False(t, rule.Enabled)
}

func TestProjectTemplateReviewRuleServiceDefaultsEnabled(t *testing.T) {
	repo := &fakeProjectTemplateReviewRuleRepository{
		templateExists: func(ctx context.Context, templateID uint) (bool, error) {
			return true, nil
		},
		create: func(ctx context.Context, input ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error) {
			require.True(t, input.Enabled)
			return &ProjectTemplateReviewRule{ID: 1, Enabled: input.Enabled}, nil
		},
	}
	svc := NewProjectTemplateReviewRuleService(repo)

	_, err := svc.Create(context.Background(), ProjectTemplateReviewRuleInput{
		TemplateID:   7,
		Name:         "Rule",
		GlobPatterns: []string{"*.go"},
		Content:      "Check Go files",
	})

	require.NoError(t, err)
}

func TestProjectTemplateReviewRuleServiceRejectsInvalidCreate(t *testing.T) {
	svc := NewProjectTemplateReviewRuleService(&fakeProjectTemplateReviewRuleRepository{})

	_, err := svc.Create(context.Background(), ProjectTemplateReviewRuleInput{
		Name:         " ",
		GlobPatterns: []string{"*.go"},
		Content:      "content",
	})

	require.ErrorIs(t, err, ErrInvalidProjectTemplateReviewRuleInput)
}

func TestProjectTemplateReviewRuleServiceRejectsMissingTemplate(t *testing.T) {
	repo := &fakeProjectTemplateReviewRuleRepository{
		templateExists: func(ctx context.Context, templateID uint) (bool, error) {
			require.Equal(t, uint(404), templateID)
			return false, nil
		},
	}
	svc := NewProjectTemplateReviewRuleService(repo)

	_, err := svc.Create(context.Background(), ProjectTemplateReviewRuleInput{
		TemplateID:   404,
		Name:         "Rule",
		GlobPatterns: []string{"*.go"},
		Content:      "content",
	})

	require.ErrorIs(t, err, ErrProjectTemplateNotFound)
}

func TestProjectTemplateReviewRuleServiceUpdateRequiresRuleToBelongToTemplate(t *testing.T) {
	repo := &fakeProjectTemplateReviewRuleRepository{
		templateExists: func(ctx context.Context, templateID uint) (bool, error) {
			return true, nil
		},
		find: func(ctx context.Context, id uint) (*ProjectTemplateReviewRule, error) {
			require.Equal(t, uint(9), id)
			return &ProjectTemplateReviewRule{ID: 9, TemplateID: 7}, nil
		},
	}
	svc := NewProjectTemplateReviewRuleService(repo)

	_, err := svc.Update(context.Background(), 9, ProjectTemplateReviewRuleInput{
		TemplateID:   8,
		Name:         "Rule",
		GlobPatterns: []string{"*.go"},
		Content:      "content",
	})

	require.ErrorIs(t, err, ErrProjectTemplateReviewRuleTemplateMismatch)
}

func TestProjectTemplateReviewRuleServiceListRequiresTemplateID(t *testing.T) {
	svc := NewProjectTemplateReviewRuleService(&fakeProjectTemplateReviewRuleRepository{})

	_, err := svc.ListByTemplateID(context.Background(), 0)

	require.ErrorIs(t, err, ErrInvalidProjectTemplateReviewRuleInput)
}

func TestProjectTemplateReviewRuleServiceGetAndDeleteRequireID(t *testing.T) {
	svc := NewProjectTemplateReviewRuleService(&fakeProjectTemplateReviewRuleRepository{})

	_, err := svc.Get(context.Background(), 0)
	require.ErrorIs(t, err, ErrInvalidProjectTemplateReviewRuleInput)

	err = svc.Delete(context.Background(), 0)
	require.ErrorIs(t, err, ErrInvalidProjectTemplateReviewRuleInput)
}

type fakeProjectTemplateReviewRuleRepository struct {
	create         func(context.Context, ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error)
	update         func(context.Context, uint, ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error)
	find           func(context.Context, uint) (*ProjectTemplateReviewRule, error)
	delete         func(context.Context, uint) error
	listByTemplate func(context.Context, uint) ([]ProjectTemplateReviewRule, error)
	templateExists func(context.Context, uint) (bool, error)
}

func (r *fakeProjectTemplateReviewRuleRepository) CreateProjectTemplateReviewRule(ctx context.Context, input ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error) {
	if r.create == nil {
		return nil, errors.New("unexpected create")
	}
	return r.create(ctx, input)
}

func (r *fakeProjectTemplateReviewRuleRepository) UpdateProjectTemplateReviewRule(ctx context.Context, id uint, input ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error) {
	if r.update == nil {
		return nil, errors.New("unexpected update")
	}
	return r.update(ctx, id, input)
}

func (r *fakeProjectTemplateReviewRuleRepository) FindProjectTemplateReviewRuleByID(ctx context.Context, id uint) (*ProjectTemplateReviewRule, error) {
	if r.find == nil {
		return nil, errors.New("unexpected find")
	}
	return r.find(ctx, id)
}

func (r *fakeProjectTemplateReviewRuleRepository) DeleteProjectTemplateReviewRule(ctx context.Context, id uint) error {
	if r.delete == nil {
		return errors.New("unexpected delete")
	}
	return r.delete(ctx, id)
}

func (r *fakeProjectTemplateReviewRuleRepository) ListProjectTemplateReviewRulesByTemplateID(ctx context.Context, templateID uint) ([]ProjectTemplateReviewRule, error) {
	if r.listByTemplate == nil {
		return nil, errors.New("unexpected list by template")
	}
	return r.listByTemplate(ctx, templateID)
}

func (r *fakeProjectTemplateReviewRuleRepository) ProjectTemplateExists(ctx context.Context, templateID uint) (bool, error) {
	if r.templateExists == nil {
		return false, errors.New("unexpected template exists")
	}
	return r.templateExists(ctx, templateID)
}
