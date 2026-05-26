package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnalysisExecutionLogServiceGetValidatesID(t *testing.T) {
	svc := NewAnalysisExecutionLogService(&fakeAnalysisExecutionLogRepository{})

	got, err := svc.Get(context.Background(), 0)

	require.Nil(t, got)
	require.ErrorIs(t, err, ErrInvalidReviewLogInput)
}

func TestAnalysisExecutionLogServiceSearchNormalizesQuery(t *testing.T) {
	repo := &fakeAnalysisExecutionLogRepository{
		search: func(ctx context.Context, query AnalysisExecutionLogSearchQuery) (*AnalysisExecutionLogPage, error) {
			require.Equal(t, uint(7), query.ProjectID)
			require.Equal(t, uint(3), query.PlanID)
			require.Equal(t, "succeeded", query.Status)
			require.Equal(t, 1, query.Page)
			require.Equal(t, 200, query.Size)
			return &AnalysisExecutionLogPage{
				Items: []ProjectAnalysisPlanExecutionLog{{ID: 11, Status: "succeeded"}},
				Total: 1,
				Page:  1,
				Size:  200,
			}, nil
		},
	}
	svc := NewAnalysisExecutionLogService(repo)

	page, err := svc.Search(context.Background(), AnalysisExecutionLogSearchQuery{
		ProjectID: 7,
		PlanID:    3,
		Status:    " succeeded ",
		Page:      -1,
		Size:      500,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
}

func TestAnalysisExecutionLogServiceGenerateShareToken(t *testing.T) {
	repo := &fakeAnalysisExecutionLogRepository{
		updateShareToken: func(ctx context.Context, id uint, token string, expiresAt int64) (*ProjectAnalysisPlanExecutionLog, error) {
			require.Equal(t, uint(9), id)
			require.Len(t, token, 32)
			require.Greater(t, expiresAt, int64(0))
			return &ProjectAnalysisPlanExecutionLog{
				ID:                  id,
				ShareToken:          token,
				ShareTokenExpiresAt: expiresAt,
			}, nil
		},
	}
	svc := NewAnalysisExecutionLogService(repo)

	token, err := svc.GenerateShareToken(context.Background(), 9)

	require.NoError(t, err)
	require.Len(t, token.ShareToken, 32)
	require.Greater(t, token.ShareTokenExpiresAt, int64(0))
}

type fakeAnalysisExecutionLogRepository struct {
	find             func(context.Context, uint) (*ProjectAnalysisPlanExecutionLog, error)
	search           func(context.Context, AnalysisExecutionLogSearchQuery) (*AnalysisExecutionLogPage, error)
	updateShareToken func(context.Context, uint, string, int64) (*ProjectAnalysisPlanExecutionLog, error)
}

func (r *fakeAnalysisExecutionLogRepository) FindAnalysisExecutionByID(ctx context.Context, id uint) (*ProjectAnalysisPlanExecutionLog, error) {
	if r.find != nil {
		return r.find(ctx, id)
	}
	return nil, ErrReviewLogNotFound
}

func (r *fakeAnalysisExecutionLogRepository) SearchAnalysisExecution(ctx context.Context, query AnalysisExecutionLogSearchQuery) (*AnalysisExecutionLogPage, error) {
	return r.search(ctx, query)
}

func (r *fakeAnalysisExecutionLogRepository) UpdateAnalysisExecutionShareToken(ctx context.Context, id uint, token string, expiresAt int64) (*ProjectAnalysisPlanExecutionLog, error) {
	return r.updateShareToken(ctx, id, token, expiresAt)
}
