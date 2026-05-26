package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIMRobotServiceCreateNormalizesInputAndDefaultsEnabled(t *testing.T) {
	repo := &fakeIMRobotRepository{
		create: func(ctx context.Context, input IMRobotInput) (*IMRobot, error) {
			require.Equal(t, IMRobotPlatformDingTalk, input.Platform)
			require.Equal(t, "review alerts", input.Name)
			require.Equal(t, "https://example.com/webhook", input.WebhookURL)
			require.Equal(t, "secret", input.Secret)
			require.True(t, input.Enabled)
			return &IMRobot{ID: 1, Name: input.Name, Enabled: input.Enabled}, nil
		},
	}
	svc := NewIMRobotService(repo)

	robot, err := svc.Create(context.Background(), IMRobotInput{
		Platform:   " dingtalk ",
		Name:       " review alerts ",
		WebhookURL: " https://example.com/webhook ",
		Secret:     " secret ",
	})

	require.NoError(t, err)
	require.Equal(t, uint(1), robot.ID)
	require.True(t, robot.Enabled)
}

func TestIMRobotServiceCreatePreservesExplicitDisabled(t *testing.T) {
	enabled := false
	repo := &fakeIMRobotRepository{
		create: func(ctx context.Context, input IMRobotInput) (*IMRobot, error) {
			require.False(t, input.Enabled)
			return &IMRobot{ID: 1, Enabled: input.Enabled}, nil
		},
	}
	svc := NewIMRobotService(repo)

	robot, err := svc.Create(context.Background(), IMRobotInput{
		Platform:   IMRobotPlatformFeishu,
		Name:       "feishu",
		WebhookURL: "https://example.com/webhook",
		EnabledSet: true,
		Enabled:    enabled,
	})

	require.NoError(t, err)
	require.False(t, robot.Enabled)
}

func TestIMRobotServiceRejectsInvalidCreate(t *testing.T) {
	svc := NewIMRobotService(&fakeIMRobotRepository{})

	_, err := svc.Create(context.Background(), IMRobotInput{Name: "robot"})

	require.ErrorIs(t, err, ErrInvalidIMRobotInput)
}

func TestIMRobotServiceRejectsUnsafeWebhookURL(t *testing.T) {
	svc := NewIMRobotService(&fakeIMRobotRepository{})

	tests := []string{
		"/relative/path",
		"ftp://example.com/webhook",
		"mailto:ops@example.com",
		"javascript:alert(1)",
		"https:///missing-host",
	}

	for _, webhookURL := range tests {
		t.Run(webhookURL, func(t *testing.T) {
			_, err := svc.Create(context.Background(), IMRobotInput{
				Platform:   IMRobotPlatformDingTalk,
				Name:       "alerts",
				WebhookURL: webhookURL,
			})

			require.ErrorIs(t, err, ErrInvalidIMRobotInput)
		})
	}
}

func TestIMRobotServiceRejectsOverlongInput(t *testing.T) {
	svc := NewIMRobotService(&fakeIMRobotRepository{})

	tests := []IMRobotInput{
		{Platform: IMRobotPlatformDingTalk, Name: strings.Repeat("n", 129), WebhookURL: "https://example.com/webhook"},
		{Platform: IMRobotPlatformDingTalk, Name: "alerts", WebhookURL: "https://example.com/" + strings.Repeat("w", 1005)},
		{Platform: IMRobotPlatformDingTalk, Name: "alerts", WebhookURL: "https://example.com/webhook", Secret: strings.Repeat("s", 513)},
	}

	for _, input := range tests {
		_, err := svc.Create(context.Background(), input)
		require.ErrorIs(t, err, ErrInvalidIMRobotInput)
	}
}

func TestIMRobotServiceUpdateRequiresID(t *testing.T) {
	svc := NewIMRobotService(&fakeIMRobotRepository{})

	_, err := svc.Update(context.Background(), 0, IMRobotInput{Name: "robot"})

	require.ErrorIs(t, err, ErrInvalidIMRobotInput)
}

func TestIMRobotServiceGetRequiresID(t *testing.T) {
	svc := NewIMRobotService(&fakeIMRobotRepository{})

	_, err := svc.Get(context.Background(), 0)

	require.ErrorIs(t, err, ErrInvalidIMRobotInput)
}

func TestIMRobotServiceDeleteCleansIDs(t *testing.T) {
	repo := &fakeIMRobotRepository{
		countReferences: func(ctx context.Context, ids []uint) (int64, error) {
			require.Equal(t, []uint{3, 2}, ids)
			return 0, nil
		},
		delete: func(ctx context.Context, ids []uint) error {
			require.Equal(t, []uint{3, 2}, ids)
			return nil
		},
	}
	svc := NewIMRobotService(repo)

	err := svc.Delete(context.Background(), []uint{0, 3, 3, 2})

	require.NoError(t, err)
}

func TestIMRobotServiceDeleteRejectsReferencedRobots(t *testing.T) {
	repo := &fakeIMRobotRepository{
		countReferences: func(ctx context.Context, ids []uint) (int64, error) {
			require.Equal(t, []uint{3}, ids)
			return 1, nil
		},
	}
	svc := NewIMRobotService(repo)

	err := svc.Delete(context.Background(), []uint{3})

	require.ErrorIs(t, err, ErrIMRobotInUse)
}

func TestIMRobotServiceDeleteRejectsEmptyIDs(t *testing.T) {
	svc := NewIMRobotService(&fakeIMRobotRepository{})

	err := svc.Delete(context.Background(), []uint{0})

	require.ErrorIs(t, err, ErrInvalidIMRobotInput)
}

func TestIMRobotServiceSearchNormalizesQuery(t *testing.T) {
	repo := &fakeIMRobotRepository{
		search: func(ctx context.Context, query IMRobotSearchQuery) (*IMRobotPage, error) {
			require.Equal(t, "alerts", query.Keyword)
			require.Equal(t, IMRobotPlatformWeCom, query.Platform)
			require.NotNil(t, query.Enabled)
			require.True(t, *query.Enabled)
			require.Equal(t, 1, query.Page)
			require.Equal(t, 20, query.Size)
			return &IMRobotPage{Items: []IMRobot{{ID: 1, Name: "alerts"}}, Total: 1, Page: 1, Size: 20}, nil
		},
	}
	enabled := true
	svc := NewIMRobotService(repo)

	page, err := svc.Search(context.Background(), IMRobotSearchQuery{
		Keyword:  " alerts ",
		Platform: " wecom ",
		Enabled:  &enabled,
		Page:     -1,
		Size:     0,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
}

func TestIMRobotServicePropagatesNotFound(t *testing.T) {
	svc := NewIMRobotService(&fakeIMRobotRepository{
		find: func(ctx context.Context, id uint) (*IMRobot, error) {
			return nil, ErrIMRobotNotFound
		},
	})

	_, err := svc.Get(context.Background(), 404)

	require.ErrorIs(t, err, ErrIMRobotNotFound)
}

type fakeIMRobotRepository struct {
	create          func(context.Context, IMRobotInput) (*IMRobot, error)
	update          func(context.Context, uint, IMRobotInput) (*IMRobot, error)
	find            func(context.Context, uint) (*IMRobot, error)
	delete          func(context.Context, []uint) error
	search          func(context.Context, IMRobotSearchQuery) (*IMRobotPage, error)
	listEnabled     func(context.Context) ([]IMRobot, error)
	countReferences func(context.Context, []uint) (int64, error)
}

func (r *fakeIMRobotRepository) CreateIMRobot(ctx context.Context, input IMRobotInput) (*IMRobot, error) {
	if r.create == nil {
		return nil, errors.New("unexpected create")
	}
	return r.create(ctx, input)
}

func (r *fakeIMRobotRepository) UpdateIMRobot(ctx context.Context, id uint, input IMRobotInput) (*IMRobot, error) {
	if r.update == nil {
		return nil, errors.New("unexpected update")
	}
	return r.update(ctx, id, input)
}

func (r *fakeIMRobotRepository) FindIMRobotByID(ctx context.Context, id uint) (*IMRobot, error) {
	if r.find == nil {
		return nil, errors.New("unexpected find")
	}
	return r.find(ctx, id)
}

func (r *fakeIMRobotRepository) DeleteIMRobots(ctx context.Context, ids []uint) error {
	if r.delete == nil {
		return errors.New("unexpected delete")
	}
	return r.delete(ctx, ids)
}

func (r *fakeIMRobotRepository) SearchIMRobots(ctx context.Context, query IMRobotSearchQuery) (*IMRobotPage, error) {
	if r.search == nil {
		return nil, errors.New("unexpected search")
	}
	return r.search(ctx, query)
}

func (r *fakeIMRobotRepository) ListEnabledIMRobots(ctx context.Context) ([]IMRobot, error) {
	if r.listEnabled == nil {
		return nil, errors.New("unexpected list enabled")
	}
	return r.listEnabled(ctx)
}

func (r *fakeIMRobotRepository) CountIMRobotReferences(ctx context.Context, ids []uint) (int64, error) {
	if r.countReferences == nil {
		return 0, errors.New("unexpected count references")
	}
	return r.countReferences(ctx, ids)
}
