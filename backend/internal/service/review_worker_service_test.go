package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReviewWorkerProcessNextReturnsFalseWhenNoTask(t *testing.T) {
	tasks := &fakeWorkerTaskRunner{claimErr: ErrReviewTaskNotFound}
	worker := NewReviewWorkerService(tasks, &fakeWorkerProjectRepo{}, &fakeWorkerLLMModelRepo{}, &fakeWorkerGitLabClient{}, &fakeWorkerLLMClient{})

	result, err := worker.ProcessNext(context.Background(), "worker-1")

	require.NoError(t, err)
	require.False(t, result.Processed)
}

func TestReviewWorkerProcessNextHandlesPushTask(t *testing.T) {
	tasks := &fakeWorkerTaskRunner{
		task: &ReviewTask{
			ID:          10,
			ProjectID:   7,
			EventType:   ReviewTaskEventPush,
			PayloadJSON: []byte(`{"project":{"id":123,"web_url":"https://gitlab.example.com/group/repo"},"after":"abc123"}`),
		},
	}
	projects := &fakeWorkerProjectRepo{
		project: &Project{
			ID:          7,
			Name:        "repo",
			WebURL:      "https://gitlab.example.com/group/repo",
			AccessToken: "project-token",
		},
	}
	models := &fakeWorkerLLMModelRepo{
		model: &LLMModel{
			ID:         3,
			Provider:   "openai",
			ModelCode:  "gpt-test",
			APIBaseURL: "https://llm.example.com/v1",
			APIKey:     "llm-key",
			MaxTokens:  4096,
			IsDefault:  true,
		},
	}
	gitlab := &fakeWorkerGitLabClient{
		commitDiff: []GitLabDiff{{OldPath: "main.go", NewPath: "main.go", Diff: "@@ diff"}},
	}
	llm := &fakeWorkerLLMClient{response: "review ok"}
	worker := NewReviewWorkerService(tasks, projects, models, gitlab, llm)

	result, err := worker.ProcessNext(context.Background(), "worker-1")

	require.NoError(t, err)
	require.True(t, result.Processed)
	require.Equal(t, uint(10), result.TaskID)
	require.Equal(t, "review ok", result.ReviewText)
	require.Equal(t, "https://gitlab.example.com", gitlab.lastBaseURL)
	require.Equal(t, "project-token", gitlab.lastToken)
	require.Equal(t, 123, gitlab.lastCommitProjectID)
	require.Equal(t, "abc123", gitlab.lastCommitSHA)
	require.Contains(t, llm.lastPrompt, "@@ diff")
	require.Equal(t, "gpt-test", llm.lastInput.ModelCode)
	require.True(t, tasks.started)
	require.True(t, tasks.succeeded)
}

func TestReviewWorkerProcessNextHandlesMergeRequestTask(t *testing.T) {
	tasks := &fakeWorkerTaskRunner{
		task: &ReviewTask{
			ID:          11,
			ProjectID:   8,
			EventType:   ReviewTaskEventMergeRequest,
			PayloadJSON: []byte(`{"project":{"id":456,"web_url":"https://gitlab.example.com/group/repo"},"object_attributes":{"iid":5,"last_commit":{"id":"def456"}}}`),
		},
	}
	projects := &fakeWorkerProjectRepo{project: &Project{ID: 8, WebURL: "https://gitlab.example.com/group/repo", AccessToken: "project-token"}}
	models := &fakeWorkerLLMModelRepo{model: &LLMModel{ModelCode: "gpt-test", APIBaseURL: "https://llm.example.com/v1", APIKey: "llm-key"}}
	gitlab := &fakeWorkerGitLabClient{mrDiff: []GitLabDiff{{OldPath: "a.go", NewPath: "a.go", Diff: "@@ mr"}}}
	llm := &fakeWorkerLLMClient{response: "mr review ok"}
	worker := NewReviewWorkerService(tasks, projects, models, gitlab, llm)

	result, err := worker.ProcessNext(context.Background(), "worker-1")

	require.NoError(t, err)
	require.True(t, result.Processed)
	require.Equal(t, 456, gitlab.lastMRProjectID)
	require.Equal(t, 5, gitlab.lastMRIID)
	require.Contains(t, llm.lastPrompt, "@@ mr")
	require.True(t, tasks.succeeded)
}

func TestReviewWorkerProcessNextMarksFailedWhenLLMFails(t *testing.T) {
	llmErr := errors.New("llm unavailable")
	tasks := &fakeWorkerTaskRunner{
		task: &ReviewTask{
			ID:          12,
			ProjectID:   9,
			EventType:   ReviewTaskEventPush,
			PayloadJSON: []byte(`{"project":{"id":789,"web_url":"https://gitlab.example.com/group/repo"},"after":"abc123"}`),
		},
	}
	worker := NewReviewWorkerService(
		tasks,
		&fakeWorkerProjectRepo{project: &Project{ID: 9, WebURL: "https://gitlab.example.com/group/repo", AccessToken: "project-token"}},
		&fakeWorkerLLMModelRepo{model: &LLMModel{ModelCode: "gpt-test", APIBaseURL: "https://llm.example.com/v1", APIKey: "llm-key"}},
		&fakeWorkerGitLabClient{commitDiff: []GitLabDiff{{NewPath: "main.go", Diff: "@@ diff"}}},
		&fakeWorkerLLMClient{err: llmErr},
	)

	result, err := worker.ProcessNext(context.Background(), "worker-1")

	require.ErrorIs(t, err, llmErr)
	require.True(t, result.Processed)
	require.True(t, tasks.failed)
	require.Contains(t, tasks.failedMessage, "llm unavailable")
	require.False(t, tasks.succeeded)
}

type fakeWorkerTaskRunner struct {
	task          *ReviewTask
	claimErr      error
	started       bool
	succeeded     bool
	failed        bool
	failedMessage string
}

func (r *fakeWorkerTaskRunner) ClaimNext(ctx context.Context, workerID string) (*ReviewTask, error) {
	if r.claimErr != nil {
		return nil, r.claimErr
	}
	return cloneReviewTask(r.task), nil
}

func (r *fakeWorkerTaskRunner) StartAttempt(ctx context.Context, taskID uint) (*ReviewTaskAttempt, error) {
	r.started = true
	return &ReviewTaskAttempt{ID: 1, TaskID: taskID, AttemptNo: 1, StartedAt: time.Now()}, nil
}

func (r *fakeWorkerTaskRunner) MarkSucceeded(ctx context.Context, taskID uint) error {
	r.succeeded = true
	return nil
}

func (r *fakeWorkerTaskRunner) MarkFailed(ctx context.Context, taskID uint, message string) error {
	r.failed = true
	r.failedMessage = message
	return nil
}

type fakeWorkerProjectRepo struct {
	project *Project
}

func (r *fakeWorkerProjectRepo) FindByID(ctx context.Context, id uint) (*Project, error) {
	if r.project == nil {
		return nil, ErrProjectNotFound
	}
	return cloneProject(r.project), nil
}

type fakeWorkerLLMModelRepo struct {
	model *LLMModel
}

func (r *fakeWorkerLLMModelRepo) Default(ctx context.Context) (*LLMModel, error) {
	if r.model == nil {
		return nil, ErrLLMModelNotFound
	}
	return cloneLLMModel(r.model), nil
}

type fakeWorkerGitLabClient struct {
	lastBaseURL         string
	lastToken           string
	lastCommitProjectID int
	lastCommitSHA       string
	lastMRProjectID     int
	lastMRIID           int
	commitDiff          []GitLabDiff
	mrDiff              []GitLabDiff
}

func (c *fakeWorkerGitLabClient) WithAuth(baseURL string, token string) GitLabClient {
	c.lastBaseURL = baseURL
	c.lastToken = token
	return c
}

func (c *fakeWorkerGitLabClient) SearchProjects(ctx context.Context, opts GitLabSearchOptions) ([]GitLabProject, error) {
	return nil, nil
}

func (c *fakeWorkerGitLabClient) SearchGroups(ctx context.Context, opts GitLabSearchOptions) ([]GitLabGroup, error) {
	return nil, nil
}

func (c *fakeWorkerGitLabClient) GetMergeRequestChanges(ctx context.Context, projectID int, mergeRequestIID int) ([]GitLabDiff, error) {
	c.lastMRProjectID = projectID
	c.lastMRIID = mergeRequestIID
	return c.mrDiff, nil
}

func (c *fakeWorkerGitLabClient) GetCommitDiff(ctx context.Context, projectID int, sha string) ([]GitLabDiff, error) {
	c.lastCommitProjectID = projectID
	c.lastCommitSHA = sha
	return c.commitDiff, nil
}

type fakeWorkerLLMClient struct {
	response   string
	err        error
	lastInput  LLMChatInput
	lastPrompt string
}

func (c *fakeWorkerLLMClient) Chat(ctx context.Context, input LLMChatInput) (string, error) {
	c.lastInput = input
	if len(input.Messages) > 0 {
		c.lastPrompt = input.Messages[len(input.Messages)-1].Content
	}
	if c.err != nil {
		return "", c.err
	}
	return c.response, nil
}
