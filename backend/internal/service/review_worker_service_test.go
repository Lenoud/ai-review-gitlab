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
	worker := NewReviewWorkerService(tasks, &fakeWorkerProjectRepo{}, &fakeWorkerLLMModelRepo{}, &fakeWorkerGitLabClient{}, &fakeWorkerLLMClient{}, nil, nil)

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
			PayloadJSON: []byte(`{"project":{"id":123,"web_url":"https://gitlab.example.com/group/repo"},"user_username":"alice","user_name":"Alice","ref":"refs/heads/main","after":"abc123","commits":[{"id":"abc123","message":"fix auth\n","author":{"name":"Alice"},"url":"https://gitlab.example.com/group/repo/-/commit/abc123","timestamp":"2026-05-25T10:00:00Z"}]}`),
		},
	}
	projects := &fakeWorkerProjectRepo{
		project: &Project{
			ID:                   7,
			Name:                 "repo",
			WebURL:               "https://gitlab.example.com/group/repo",
			AccessToken:          "project-token",
			ReviewPromptTemplate: "项目 {{projectName}} 请检查。",
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
	llm := &fakeWorkerLLMClient{response: "<think>hidden reasoning</think>\n\nreview ok"}
	logs := &fakeWorkerReviewLogWriter{}
	traces := &fakeWorkerTraceWriter{}
	worker := NewReviewWorkerService(tasks, projects, models, gitlab, llm, logs, traces)

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
	require.Contains(t, llm.lastPrompt, "项目 repo 请检查。")
	require.Contains(t, llm.lastPrompt, "### 待审查内容")
	require.Contains(t, llm.lastPrompt, "fix auth (by Alice);")
	require.Equal(t, "gpt-test", llm.lastInput.ModelCode)
	require.True(t, tasks.started)
	require.True(t, tasks.succeeded)
	require.NotNil(t, logs.lastPush)
	require.Equal(t, uint(7), logs.lastPush.ProjectID)
	require.Equal(t, "repo", logs.lastPush.ProjectName)
	require.Equal(t, "main", logs.lastPush.Branch)
	require.Equal(t, "alice", logs.lastPush.AuthorIdentity)
	require.Equal(t, "Alice", logs.lastPush.AuthorDisplayName)
	require.Equal(t, "fix auth (by Alice);", logs.lastPush.CommitMessages)
	require.Equal(t, "https://gitlab.example.com/group/repo/-/commit/abc123", logs.lastPush.LastCommitURL)
	require.Equal(t, "review ok", logs.lastPush.ReviewResult)
	require.NotNil(t, traces.last)
	require.Equal(t, "push", traces.last.ReviewEventType)
	require.Equal(t, uint(101), traces.last.ReviewEventID)
	require.Contains(t, traces.last.Prompt, "@@ diff")
	require.Equal(t, "review ok", traces.last.Response)
	require.Equal(t, "openai", traces.last.Provider)
	require.Equal(t, "gpt-test", traces.last.ModelCode)
}

func TestReviewWorkerProcessNextHandlesMergeRequestTask(t *testing.T) {
	tasks := &fakeWorkerTaskRunner{
		task: &ReviewTask{
			ID:          11,
			ProjectID:   8,
			EventType:   ReviewTaskEventMergeRequest,
			PayloadJSON: []byte(`{"project":{"id":456,"web_url":"https://gitlab.example.com/group/repo"},"user":{"username":"bob","name":"Bob"},"object_attributes":{"iid":5,"source_branch":"feature/login","target_branch":"main","url":"https://gitlab.example.com/group/repo/-/merge_requests/5","last_commit":{"id":"def456","message":"add login"}}}`),
		},
	}
	projects := &fakeWorkerProjectRepo{project: &Project{ID: 8, Name: "repo", WebURL: "https://gitlab.example.com/group/repo", AccessToken: "project-token"}}
	models := &fakeWorkerLLMModelRepo{model: &LLMModel{ModelCode: "gpt-test", APIBaseURL: "https://llm.example.com/v1", APIKey: "llm-key"}}
	gitlab := &fakeWorkerGitLabClient{mrDiff: []GitLabDiff{{OldPath: "a.go", NewPath: "a.go", Diff: "@@ mr"}}}
	llm := &fakeWorkerLLMClient{response: "mr review ok"}
	logs := &fakeWorkerReviewLogWriter{}
	traces := &fakeWorkerTraceWriter{}
	worker := NewReviewWorkerService(tasks, projects, models, gitlab, llm, logs, traces)

	result, err := worker.ProcessNext(context.Background(), "worker-1")

	require.NoError(t, err)
	require.True(t, result.Processed)
	require.Equal(t, 456, gitlab.lastMRProjectID)
	require.Equal(t, 5, gitlab.lastMRIID)
	require.Contains(t, llm.lastPrompt, "@@ mr")
	require.True(t, tasks.succeeded)
	require.NotNil(t, logs.lastMerge)
	require.Equal(t, uint(8), logs.lastMerge.ProjectID)
	require.Equal(t, "feature/login", logs.lastMerge.SourceBranch)
	require.Equal(t, "main", logs.lastMerge.TargetBranch)
	require.Equal(t, "def456", logs.lastMerge.LastCommitID)
	require.Equal(t, "mr review ok", logs.lastMerge.ReviewResult)
	require.NotNil(t, traces.last)
	require.Equal(t, "merge_request", traces.last.ReviewEventType)
	require.Equal(t, uint(202), traces.last.ReviewEventID)
	require.Contains(t, traces.last.Prompt, "@@ mr")
	require.Equal(t, "mr review ok", traces.last.Response)
	require.Equal(t, "gpt-test", traces.last.ModelCode)
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
		&fakeWorkerReviewLogWriter{},
		nil,
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

type fakeWorkerReviewLogWriter struct {
	lastPush  *PushReviewLogInput
	lastMerge *MergeRequestReviewLogInput
}

func (w *fakeWorkerReviewLogWriter) CreatePush(ctx context.Context, input PushReviewLogInput) (*PushReviewLog, error) {
	copied := input
	w.lastPush = &copied
	return &PushReviewLog{ID: 101}, nil
}

func (w *fakeWorkerReviewLogWriter) CreateMergeRequest(ctx context.Context, input MergeRequestReviewLogInput) (*MergeRequestReviewLog, error) {
	copied := input
	w.lastMerge = &copied
	return &MergeRequestReviewLog{ID: 202}, nil
}

type fakeWorkerTraceWriter struct {
	last *AIReviewTraceInput
}

func (w *fakeWorkerTraceWriter) Create(ctx context.Context, input AIReviewTraceInput) (*AIReviewTrace, error) {
	copied := input
	w.last = &copied
	return &AIReviewTrace{ID: 303}, nil
}
