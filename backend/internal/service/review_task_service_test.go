package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReviewTaskServiceCreatesPushTaskAndDedupes(t *testing.T) {
	projects := &memoryReviewProjectRepository{
		byWebURL: map[string]*Project{
			"https://gitlab.example.com/group/ai-review": {
				ID:     1,
				Name:   "AI Review",
				WebURL: "https://gitlab.example.com/group/ai-review",
			},
		},
	}
	tasks := newMemoryReviewTaskRepository()
	svc := NewReviewTaskService(projects, tasks, ReviewTaskOptions{
		Now: func() time.Time { return time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC) },
	})

	result, err := svc.EnqueueGitLabWebhook(context.Background(), GitLabWebhookInput{
		Event:   "Push Hook",
		Payload: []byte(`{"project":{"web_url":"https://gitlab.example.com/group/ai-review"},"ref":"refs/heads/main","after":"abc123"}`),
	})
	require.NoError(t, err)
	require.False(t, result.Duplicate)
	require.Equal(t, ReviewTaskStatusPending, result.Status)

	again, err := svc.EnqueueGitLabWebhook(context.Background(), GitLabWebhookInput{
		Event:   "Push Hook",
		Payload: []byte(`{"project":{"web_url":"https://gitlab.example.com/group/ai-review"},"ref":"refs/heads/main","after":"abc123"}`),
	})
	require.NoError(t, err)
	require.True(t, again.Duplicate)
	require.Equal(t, result.TaskID, again.TaskID)
	require.Equal(t, "gitlab:push:1:refs/heads/main:abc123", tasks.created[0].DedupeKey)
}

func TestReviewTaskServiceCreatesMergeRequestTask(t *testing.T) {
	projects := &memoryReviewProjectRepository{
		byWebURL: map[string]*Project{
			"https://gitlab.example.com/group/ai-review": {ID: 7, Name: "AI Review", WebURL: "https://gitlab.example.com/group/ai-review"},
		},
	}
	tasks := newMemoryReviewTaskRepository()
	svc := NewReviewTaskService(projects, tasks, ReviewTaskOptions{})

	result, err := svc.EnqueueGitLabWebhook(context.Background(), GitLabWebhookInput{
		Event:   "Merge Request Hook",
		Payload: []byte(`{"project":{"web_url":"https://gitlab.example.com/group/ai-review"},"object_attributes":{"iid":12,"action":"open","last_commit":{"id":"def456"}}}`),
	})

	require.NoError(t, err)
	require.False(t, result.Duplicate)
	require.Equal(t, ReviewTaskEventMergeRequest, tasks.created[0].EventType)
	require.Equal(t, "gitlab:mr:7:12:def456:open", tasks.created[0].DedupeKey)
}

func TestReviewTaskServiceRejectsUnknownProject(t *testing.T) {
	svc := NewReviewTaskService(&memoryReviewProjectRepository{byWebURL: map[string]*Project{}}, newMemoryReviewTaskRepository(), ReviewTaskOptions{})

	_, err := svc.EnqueueGitLabWebhook(context.Background(), GitLabWebhookInput{
		Event:   "Push Hook",
		Payload: []byte(`{"project":{"web_url":"https://gitlab.example.com/group/missing"},"ref":"refs/heads/main","after":"abc123"}`),
	})

	require.ErrorIs(t, err, ErrReviewTaskProjectNotFound)
}

func TestReviewTaskServiceMarkFailedSchedulesRetryAndFinalFailure(t *testing.T) {
	tasks := newMemoryReviewTaskRepository()
	task, _, err := tasks.CreateOrGetByDedupeKey(context.Background(), ReviewTaskCreateInput{
		ProjectID:   1,
		EventType:   ReviewTaskEventPush,
		DedupeKey:   "dedupe",
		PayloadJSON: []byte(`{}`),
		MaxAttempts: 2,
	})
	require.NoError(t, err)
	tasks.tasks[task.ID].Status = ReviewTaskStatusRunning
	svc := NewReviewTaskService(&memoryReviewProjectRepository{}, tasks, ReviewTaskOptions{
		Now: func() time.Time { return time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC) },
	})

	err = svc.MarkFailed(context.Background(), task.ID, "temporary")
	require.NoError(t, err)
	require.Equal(t, ReviewTaskStatusPending, tasks.tasks[task.ID].Status)
	require.Equal(t, 1, tasks.tasks[task.ID].Attempts)
	require.True(t, tasks.tasks[task.ID].NextRunAt.After(time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)))

	err = svc.MarkFailed(context.Background(), task.ID, "final")
	require.NoError(t, err)
	require.Equal(t, ReviewTaskStatusFailed, tasks.tasks[task.ID].Status)
	require.Equal(t, 2, tasks.tasks[task.ID].Attempts)
}

type memoryReviewProjectRepository struct {
	byWebURL map[string]*Project
}

func (r *memoryReviewProjectRepository) FindByWebURL(ctx context.Context, webURL string) (*Project, error) {
	project, ok := r.byWebURL[webURL]
	if !ok {
		return nil, ErrProjectNotFound
	}
	return cloneProject(project), nil
}

type memoryReviewTaskRepository struct {
	tasks    map[uint]*ReviewTask
	created  []ReviewTaskCreateInput
	attempts map[uint][]*ReviewTaskAttempt
	nextID   uint
}

func newMemoryReviewTaskRepository() *memoryReviewTaskRepository {
	return &memoryReviewTaskRepository{
		tasks:    map[uint]*ReviewTask{},
		attempts: map[uint][]*ReviewTaskAttempt{},
		nextID:   1,
	}
}

func (r *memoryReviewTaskRepository) CreateOrGetByDedupeKey(ctx context.Context, input ReviewTaskCreateInput) (*ReviewTask, bool, error) {
	for _, task := range r.tasks {
		if task.DedupeKey == input.DedupeKey {
			return cloneReviewTask(task), true, nil
		}
	}
	task := reviewTaskFromCreateInput(input)
	task.ID = r.nextID
	r.nextID++
	r.tasks[task.ID] = task
	r.created = append(r.created, input)
	return cloneReviewTask(task), false, nil
}

func (r *memoryReviewTaskRepository) ClaimNext(ctx context.Context, workerID string, now time.Time) (*ReviewTask, error) {
	for _, task := range r.tasks {
		if task.Status == ReviewTaskStatusPending && !task.NextRunAt.After(now) {
			task.Status = ReviewTaskStatusRunning
			task.LockedBy = workerID
			task.LockedAt = &now
			return cloneReviewTask(task), nil
		}
	}
	return nil, ErrReviewTaskNotFound
}

func (r *memoryReviewTaskRepository) StartAttempt(ctx context.Context, taskID uint, now time.Time) (*ReviewTaskAttempt, error) {
	task, ok := r.tasks[taskID]
	if !ok {
		return nil, ErrReviewTaskNotFound
	}
	attempt := &ReviewTaskAttempt{ID: uint(len(r.attempts[taskID]) + 1), TaskID: taskID, AttemptNo: task.Attempts + 1, Status: ReviewTaskAttemptStatusRunning, StartedAt: now}
	r.attempts[taskID] = append(r.attempts[taskID], attempt)
	return cloneReviewTaskAttempt(attempt), nil
}

func (r *memoryReviewTaskRepository) MarkSucceeded(ctx context.Context, taskID uint, finishedAt time.Time) error {
	task, ok := r.tasks[taskID]
	if !ok {
		return ErrReviewTaskNotFound
	}
	task.Status = ReviewTaskStatusSucceeded
	task.FinishedAt = &finishedAt
	return nil
}

func (r *memoryReviewTaskRepository) MarkFailed(ctx context.Context, taskID uint, attempts int, status string, nextRunAt *time.Time, errorMessage string, finishedAt *time.Time) error {
	task, ok := r.tasks[taskID]
	if !ok {
		return ErrReviewTaskNotFound
	}
	task.Attempts = attempts
	task.Status = status
	task.NextRunAt = time.Time{}
	if nextRunAt != nil {
		task.NextRunAt = *nextRunAt
	}
	task.ErrorMessage = errorMessage
	task.FinishedAt = finishedAt
	return nil
}

func (r *memoryReviewTaskRepository) FindByID(ctx context.Context, id uint) (*ReviewTask, error) {
	task, ok := r.tasks[id]
	if !ok {
		return nil, ErrReviewTaskNotFound
	}
	return cloneReviewTask(task), nil
}
