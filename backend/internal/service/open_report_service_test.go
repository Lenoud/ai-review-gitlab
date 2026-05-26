package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenReportServiceBuildsPushHTMLWhenTokenValid(t *testing.T) {
	logs := &fakeOpenReportRepository{
		push: &PushReviewLog{
			ID:                  12,
			ProjectName:         "ai-review",
			AuthorDisplayText:   "alice（Alice）",
			Branch:              "main",
			Score:               86,
			Additions:           12,
			Deletions:           3,
			ReviewResult:        "总分：86分\n<script>alert(1)</script>",
			ShareToken:          "token",
			ShareTokenExpiresAt: time.Now().Add(time.Hour).UnixMilli(),
		},
	}
	svc := NewOpenReportService(logs)

	html, err := svc.CodeReviewReport(context.Background(), CodeReviewReportInput{LogID: 12, LogType: "push", Token: "token"})

	require.NoError(t, err)
	require.Contains(t, html, "ai-review")
	require.Contains(t, html, "总分：86分")
	require.NotContains(t, html, "<script>")
	require.Contains(t, html, "&lt;script&gt;alert(1)&lt;/script&gt;")
}

func TestOpenReportServiceRejectsExpiredOrMismatchedToken(t *testing.T) {
	svc := NewOpenReportService(&fakeOpenReportRepository{
		merge: &MergeRequestReviewLog{
			ID:                  13,
			ProjectName:         "ai-review",
			SourceBranch:        "feature/login",
			TargetBranch:        "main",
			ReviewResult:        "ok",
			ShareToken:          "token",
			ShareTokenExpiresAt: time.Now().Add(-time.Hour).UnixMilli(),
		},
	})

	_, err := svc.CodeReviewReport(context.Background(), CodeReviewReportInput{LogID: 13, LogType: "mr", Token: "token"})
	require.ErrorIs(t, err, ErrReviewLogNotFound)

	_, err = svc.CodeReviewReport(context.Background(), CodeReviewReportInput{LogID: 13, LogType: "mr", Token: "bad"})
	require.ErrorIs(t, err, ErrReviewLogNotFound)
}

func TestOpenReportServiceBuildsAnalysisHTMLWhenTokenValid(t *testing.T) {
	logs := &fakeOpenReportRepository{
		analysis: &ProjectAnalysisPlanExecutionLog{
			ID:                  21,
			ProjectID:           7,
			PlanID:              3,
			Status:              "succeeded",
			ResultContent:       "整体趋势稳定\n<script>alert(1)</script>",
			ShareToken:          "analysis-token",
			ShareTokenExpiresAt: time.Now().Add(time.Hour).UnixMilli(),
			StartedAt:           time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
			CompletedAt:         time.Date(2026, 5, 26, 10, 2, 0, 0, time.UTC),
			DurationMs:          120000,
		},
	}
	svc := NewOpenReportService(logs)

	html, err := svc.AnalysisReport(context.Background(), AnalysisReportInput{LogID: 21, Token: "analysis-token"})

	require.NoError(t, err)
	require.Contains(t, html, "AI Analysis Report")
	require.Contains(t, html, "整体趋势稳定")
	require.NotContains(t, html, "<script>")
	require.Contains(t, html, "&lt;script&gt;alert(1)&lt;/script&gt;")
}

type fakeOpenReportRepository struct {
	push     *PushReviewLog
	merge    *MergeRequestReviewLog
	analysis *ProjectAnalysisPlanExecutionLog
}

func (r *fakeOpenReportRepository) FindPushByID(ctx context.Context, id uint) (*PushReviewLog, error) {
	if r.push == nil || r.push.ID != id {
		return nil, ErrReviewLogNotFound
	}
	copy := *r.push
	return &copy, nil
}

func (r *fakeOpenReportRepository) FindMergeRequestByID(ctx context.Context, id uint) (*MergeRequestReviewLog, error) {
	if r.merge == nil || r.merge.ID != id {
		return nil, ErrReviewLogNotFound
	}
	copy := *r.merge
	return &copy, nil
}

func (r *fakeOpenReportRepository) FindAnalysisExecutionByID(ctx context.Context, id uint) (*ProjectAnalysisPlanExecutionLog, error) {
	if r.analysis == nil || r.analysis.ID != id {
		return nil, ErrReviewLogNotFound
	}
	copy := *r.analysis
	return &copy, nil
}
