package service

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestAnalysisExecutionLogServiceTestRunRecordsSuccess(t *testing.T) {
	startedAt := time.Date(2026, 5, 26, 11, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(2 * time.Second)
	repo := &fakeAnalysisExecutionLogRepository{
		project: &Project{ID: 7, Name: "repo"},
		model:   &LLMModel{Provider: "openai", ModelCode: "gpt-test", APIBaseURL: "https://llm.example.com/v1", APIKey: "key", MaxTokens: 2048},
		createAnalysis: func(ctx context.Context, input AnalysisExecutionLogInput) (*ProjectAnalysisPlanExecutionLog, error) {
			require.Equal(t, uint(3), input.PlanID)
			require.Equal(t, uint(7), input.ProjectID)
			require.Equal(t, AnalysisExecutionStatusSucceeded, input.Status)
			require.Equal(t, "analysis result", input.ResultContent)
			require.Empty(t, input.ErrorMessage)
			require.Equal(t, startedAt, input.StartedAt)
			require.Equal(t, completedAt, input.CompletedAt)
			require.Equal(t, int64(2000), input.DurationMs)
			return &ProjectAnalysisPlanExecutionLog{ID: 99, PlanID: input.PlanID, ProjectID: input.ProjectID, Status: input.Status, ResultContent: input.ResultContent, DurationMs: input.DurationMs}, nil
		},
	}
	llm := &fakeAnalysisLLMClient{response: "analysis result"}
	svc := NewAnalysisExecutionLogServiceWithRunner(repo, repo, repo, llm, func() time.Time {
		if repo.nowCalls == 0 {
			repo.nowCalls++
			return startedAt
		}
		return completedAt
	})

	log, err := svc.TestRun(context.Background(), AnalysisTestRunInput{
		ProjectID:         7,
		PlanID:            3,
		Prompt:            " summarize weekly risk ",
		PlanName:          " weekly ",
		HTMLReportEnabled: true,
	})

	require.NoError(t, err)
	require.Equal(t, uint(99), log.ID)
	require.Equal(t, "succeeded", log.Status)
	require.Equal(t, "analysis result", log.ResultContent)
	require.Equal(t, "gpt-test", llm.lastInput.ModelCode)
	require.Len(t, llm.lastInput.Messages, 2)
	require.Contains(t, llm.lastInput.Messages[0].Content, "项目分析助手")
	require.Contains(t, llm.lastInput.Messages[1].Content, "summarize weekly risk")
	require.Contains(t, llm.lastInput.Messages[1].Content, "repo")
	require.Contains(t, llm.lastInput.Messages[1].Content, "weekly")
}

func TestAnalysisExecutionLogServiceTestRunRecordsFailureWhenLLMFails(t *testing.T) {
	llmErr := errors.New("llm unavailable")
	repo := &fakeAnalysisExecutionLogRepository{
		project: &Project{ID: 7, Name: "repo"},
		model:   &LLMModel{ModelCode: "gpt-test", APIBaseURL: "https://llm.example.com/v1", APIKey: "key"},
		createAnalysis: func(ctx context.Context, input AnalysisExecutionLogInput) (*ProjectAnalysisPlanExecutionLog, error) {
			require.Equal(t, AnalysisExecutionStatusFailed, input.Status)
			require.Contains(t, input.ErrorMessage, "llm unavailable")
			require.Contains(t, input.ErrorStack, "llm unavailable")
			require.Empty(t, input.ResultContent)
			return &ProjectAnalysisPlanExecutionLog{ID: 100, PlanID: input.PlanID, ProjectID: input.ProjectID, Status: input.Status, ErrorMessage: input.ErrorMessage}, nil
		},
	}
	svc := NewAnalysisExecutionLogServiceWithRunner(repo, repo, repo, &fakeAnalysisLLMClient{err: llmErr}, time.Now)

	log, err := svc.TestRun(context.Background(), AnalysisTestRunInput{ProjectID: 7, Prompt: "run"})

	require.NoError(t, err)
	require.Equal(t, uint(100), log.ID)
	require.Equal(t, AnalysisExecutionStatusFailed, log.Status)
	require.Contains(t, log.ErrorMessage, "llm unavailable")
}

func TestAnalysisExecutionLogServiceTestRunRecordsFailureWhenModelMissing(t *testing.T) {
	repo := &fakeAnalysisExecutionLogRepository{
		project: &Project{ID: 7, Name: "repo"},
		createAnalysis: func(ctx context.Context, input AnalysisExecutionLogInput) (*ProjectAnalysisPlanExecutionLog, error) {
			require.Equal(t, AnalysisExecutionStatusFailed, input.Status)
			require.Contains(t, input.ErrorMessage, ErrLLMModelNotFound.Error())
			require.Equal(t, uint(7), input.ProjectID)
			return &ProjectAnalysisPlanExecutionLog{ID: 101, ProjectID: input.ProjectID, Status: input.Status, ErrorMessage: input.ErrorMessage}, nil
		},
	}
	svc := NewAnalysisExecutionLogServiceWithRunner(repo, repo, repo, &fakeAnalysisLLMClient{}, time.Now)

	log, err := svc.TestRun(context.Background(), AnalysisTestRunInput{ProjectID: 7, Prompt: "run"})

	require.NoError(t, err)
	require.Equal(t, uint(101), log.ID)
	require.Equal(t, AnalysisExecutionStatusFailed, log.Status)
	require.Contains(t, log.ErrorMessage, ErrLLMModelNotFound.Error())
}

func TestAnalysisExecutionLogServiceTestRunValidatesInput(t *testing.T) {
	svc := NewAnalysisExecutionLogServiceWithRunner(&fakeAnalysisExecutionLogRepository{}, &fakeAnalysisExecutionLogRepository{}, &fakeAnalysisExecutionLogRepository{}, &fakeAnalysisLLMClient{}, time.Now)

	_, err := svc.TestRun(context.Background(), AnalysisTestRunInput{Prompt: "missing project"})
	require.ErrorIs(t, err, ErrInvalidReviewLogInput)

	_, err = svc.TestRun(context.Background(), AnalysisTestRunInput{ProjectID: 7, Prompt: " "})
	require.ErrorIs(t, err, ErrInvalidReviewLogInput)
}

type fakeAnalysisExecutionLogRepository struct {
	find             func(context.Context, uint) (*ProjectAnalysisPlanExecutionLog, error)
	search           func(context.Context, AnalysisExecutionLogSearchQuery) (*AnalysisExecutionLogPage, error)
	updateShareToken func(context.Context, uint, string, int64) (*ProjectAnalysisPlanExecutionLog, error)
	createAnalysis   func(context.Context, AnalysisExecutionLogInput) (*ProjectAnalysisPlanExecutionLog, error)
	project          *Project
	model            *LLMModel
	nowCalls         int
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

func (r *fakeAnalysisExecutionLogRepository) CreateAnalysisExecution(ctx context.Context, input AnalysisExecutionLogInput) (*ProjectAnalysisPlanExecutionLog, error) {
	if r.createAnalysis != nil {
		return r.createAnalysis(ctx, input)
	}
	return nil, ErrInvalidReviewLogInput
}

func (r *fakeAnalysisExecutionLogRepository) FindByID(ctx context.Context, id uint) (*Project, error) {
	if r.project == nil {
		return nil, ErrProjectNotFound
	}
	return cloneProject(r.project), nil
}

func (r *fakeAnalysisExecutionLogRepository) Default(ctx context.Context) (*LLMModel, error) {
	if r.model == nil {
		return nil, ErrLLMModelNotFound
	}
	return cloneLLMModel(r.model), nil
}

type fakeAnalysisLLMClient struct {
	response  string
	err       error
	lastInput LLMChatInput
}

func (c *fakeAnalysisLLMClient) Chat(ctx context.Context, input LLMChatInput) (string, error) {
	c.lastInput = input
	if c.err != nil {
		return "", c.err
	}
	return c.response, nil
}
