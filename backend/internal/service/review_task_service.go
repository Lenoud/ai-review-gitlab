package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ReviewTaskEventPush         = "push"
	ReviewTaskEventMergeRequest = "merge_request"

	ReviewTaskStatusPending   = "pending"
	ReviewTaskStatusRunning   = "running"
	ReviewTaskStatusSucceeded = "succeeded"
	ReviewTaskStatusFailed    = "failed"
	ReviewTaskStatusCanceled  = "canceled"

	ReviewTaskAttemptStatusRunning   = "running"
	ReviewTaskAttemptStatusSucceeded = "succeeded"
	ReviewTaskAttemptStatusFailed    = "failed"
)

var (
	ErrReviewTaskNotFound        = errors.New("review task not found")
	ErrInvalidReviewTaskInput    = errors.New("invalid review task input")
	ErrUnsupportedWebhookEvent   = errors.New("unsupported webhook event")
	ErrReviewTaskProjectNotFound = errors.New("review task project not found")
)

type ReviewProjectRepository interface {
	FindByWebURL(ctx context.Context, webURL string) (*Project, error)
}

type ReviewTaskRepository interface {
	CreateOrGetByDedupeKey(ctx context.Context, input ReviewTaskCreateInput) (*ReviewTask, bool, error)
	ClaimNext(ctx context.Context, workerID string, now time.Time) (*ReviewTask, error)
	StartAttempt(ctx context.Context, taskID uint, now time.Time) (*ReviewTaskAttempt, error)
	MarkSucceeded(ctx context.Context, taskID uint, finishedAt time.Time) error
	MarkFailed(ctx context.Context, taskID uint, attempts int, status string, nextRunAt *time.Time, errorMessage string, finishedAt *time.Time) error
	FindByID(ctx context.Context, id uint) (*ReviewTask, error)
}

type ReviewTask struct {
	ID            uint       `json:"id"`
	ProjectID     uint       `json:"projectId"`
	EventType     string     `json:"eventType"`
	DedupeKey     string     `json:"dedupeKey"`
	PayloadJSON   []byte     `json:"payloadJson"`
	Status        string     `json:"status"`
	Priority      int        `json:"priority"`
	Attempts      int        `json:"attempts"`
	MaxAttempts   int        `json:"maxAttempts"`
	NextRunAt     time.Time  `json:"nextRunAt"`
	LockedBy      string     `json:"lockedBy"`
	LockedAt      *time.Time `json:"lockedAt"`
	StartedAt     *time.Time `json:"startedAt"`
	FinishedAt    *time.Time `json:"finishedAt"`
	ErrorMessage  string     `json:"errorMessage"`
	ResultLogType string     `json:"resultLogType"`
	ResultLogID   uint       `json:"resultLogId"`
}

type ReviewTaskAttempt struct {
	ID           uint       `json:"id"`
	TaskID       uint       `json:"taskId"`
	AttemptNo    int        `json:"attemptNo"`
	Status       string     `json:"status"`
	StartedAt    time.Time  `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt"`
	DurationMs   int64      `json:"durationMs"`
	ErrorMessage string     `json:"errorMessage"`
	ErrorStack   string     `json:"errorStack"`
}

type ReviewTaskCreateInput struct {
	ProjectID   uint
	EventType   string
	DedupeKey   string
	PayloadJSON []byte
	Priority    int
	MaxAttempts int
	NextRunAt   time.Time
}

type GitLabWebhookInput struct {
	Event   string
	Payload []byte
}

type ReviewTaskEnqueueResult struct {
	TaskID    uint   `json:"taskId"`
	Duplicate bool   `json:"duplicate"`
	Status    string `json:"status"`
}

type ReviewTaskOptions struct {
	Now         func() time.Time
	MaxAttempts int
}

type ReviewTaskService struct {
	projects    ReviewProjectRepository
	tasks       ReviewTaskRepository
	now         func() time.Time
	maxAttempts int
}

func NewReviewTaskService(projects ReviewProjectRepository, tasks ReviewTaskRepository, opts ReviewTaskOptions) *ReviewTaskService {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 3
	}
	return &ReviewTaskService{
		projects:    projects,
		tasks:       tasks,
		now:         opts.Now,
		maxAttempts: opts.MaxAttempts,
	}
}

func (s *ReviewTaskService) EnqueueGitLabWebhook(ctx context.Context, input GitLabWebhookInput) (*ReviewTaskEnqueueResult, error) {
	event := strings.TrimSpace(input.Event)
	if len(input.Payload) == 0 {
		return nil, ErrInvalidReviewTaskInput
	}

	parsed, err := parseGitLabWebhook(event, input.Payload)
	if err != nil {
		return nil, err
	}
	project, err := s.projects.FindByWebURL(ctx, parsed.ProjectWebURL)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, ErrReviewTaskProjectNotFound
		}
		return nil, err
	}
	dedupeKey := parsed.DedupeKey(project.ID)
	if dedupeKey == "" {
		return nil, ErrInvalidReviewTaskInput
	}

	task, duplicate, err := s.tasks.CreateOrGetByDedupeKey(ctx, ReviewTaskCreateInput{
		ProjectID:   project.ID,
		EventType:   parsed.EventType,
		DedupeKey:   dedupeKey,
		PayloadJSON: append([]byte(nil), input.Payload...),
		MaxAttempts: s.maxAttempts,
		NextRunAt:   s.now(),
	})
	if err != nil {
		return nil, err
	}
	return &ReviewTaskEnqueueResult{
		TaskID:    task.ID,
		Duplicate: duplicate,
		Status:    task.Status,
	}, nil
}

func (s *ReviewTaskService) ClaimNext(ctx context.Context, workerID string) (*ReviewTask, error) {
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, ErrInvalidReviewTaskInput
	}
	return s.tasks.ClaimNext(ctx, workerID, s.now())
}

func (s *ReviewTaskService) StartAttempt(ctx context.Context, taskID uint) (*ReviewTaskAttempt, error) {
	if taskID == 0 {
		return nil, ErrInvalidReviewTaskInput
	}
	return s.tasks.StartAttempt(ctx, taskID, s.now())
}

func (s *ReviewTaskService) MarkSucceeded(ctx context.Context, taskID uint) error {
	if taskID == 0 {
		return ErrInvalidReviewTaskInput
	}
	return s.tasks.MarkSucceeded(ctx, taskID, s.now())
}

func (s *ReviewTaskService) MarkFailed(ctx context.Context, taskID uint, message string) error {
	if taskID == 0 {
		return ErrInvalidReviewTaskInput
	}
	task, err := s.tasks.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	now := s.now()
	attempts := task.Attempts + 1
	if attempts >= task.MaxAttempts {
		return s.tasks.MarkFailed(ctx, taskID, attempts, ReviewTaskStatusFailed, nil, message, &now)
	}
	backoff := time.Duration(1<<maxInt(attempts-1, 0)) * time.Minute
	nextRunAt := now.Add(backoff)
	return s.tasks.MarkFailed(ctx, taskID, attempts, ReviewTaskStatusPending, &nextRunAt, message, nil)
}

type parsedGitLabWebhook struct {
	EventType     string
	ProjectWebURL string
	Ref           string
	AfterSHA      string
	MRIID         int
	MRAction      string
	MRLastCommit  string
}

func parseGitLabWebhook(event string, payload []byte) (*parsedGitLabWebhook, error) {
	switch event {
	case "Push Hook":
		var body struct {
			Project struct {
				WebURL string `json:"web_url"`
			} `json:"project"`
			Ref   string `json:"ref"`
			After string `json:"after"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, ErrInvalidReviewTaskInput
		}
		projectWebURL := normalizeWebURL(body.Project.WebURL)
		if projectWebURL == "" || strings.TrimSpace(body.Ref) == "" || strings.TrimSpace(body.After) == "" {
			return nil, ErrInvalidReviewTaskInput
		}
		return &parsedGitLabWebhook{
			EventType:     ReviewTaskEventPush,
			ProjectWebURL: projectWebURL,
			Ref:           strings.TrimSpace(body.Ref),
			AfterSHA:      strings.TrimSpace(body.After),
		}, nil
	case "Merge Request Hook":
		var body struct {
			Project struct {
				WebURL string `json:"web_url"`
			} `json:"project"`
			ObjectAttributes struct {
				IID        int    `json:"iid"`
				Action     string `json:"action"`
				LastCommit struct {
					ID string `json:"id"`
				} `json:"last_commit"`
			} `json:"object_attributes"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, ErrInvalidReviewTaskInput
		}
		projectWebURL := normalizeWebURL(body.Project.WebURL)
		action := strings.TrimSpace(body.ObjectAttributes.Action)
		lastCommit := strings.TrimSpace(body.ObjectAttributes.LastCommit.ID)
		if projectWebURL == "" || body.ObjectAttributes.IID == 0 || action == "" || lastCommit == "" {
			return nil, ErrInvalidReviewTaskInput
		}
		return &parsedGitLabWebhook{
			EventType:     ReviewTaskEventMergeRequest,
			ProjectWebURL: projectWebURL,
			MRIID:         body.ObjectAttributes.IID,
			MRAction:      action,
			MRLastCommit:  lastCommit,
		}, nil
	default:
		return nil, ErrUnsupportedWebhookEvent
	}
}

func (p *parsedGitLabWebhook) DedupeKey(projectID uint) string {
	switch p.EventType {
	case ReviewTaskEventPush:
		return fmt.Sprintf("gitlab:push:%d:%s:%s", projectID, p.Ref, p.AfterSHA)
	case ReviewTaskEventMergeRequest:
		return fmt.Sprintf("gitlab:mr:%d:%d:%s:%s", projectID, p.MRIID, p.MRLastCommit, p.MRAction)
	default:
		return ""
	}
}

func reviewTaskFromCreateInput(input ReviewTaskCreateInput) *ReviewTask {
	maxAttempts := input.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	return &ReviewTask{
		ProjectID:   input.ProjectID,
		EventType:   input.EventType,
		DedupeKey:   input.DedupeKey,
		PayloadJSON: append([]byte(nil), input.PayloadJSON...),
		Status:      ReviewTaskStatusPending,
		Priority:    input.Priority,
		MaxAttempts: maxAttempts,
		NextRunAt:   input.NextRunAt,
	}
}

func cloneReviewTask(task *ReviewTask) *ReviewTask {
	if task == nil {
		return nil
	}
	copy := *task
	copy.PayloadJSON = append([]byte(nil), task.PayloadJSON...)
	return &copy
}

func cloneReviewTaskAttempt(attempt *ReviewTaskAttempt) *ReviewTaskAttempt {
	if attempt == nil {
		return nil
	}
	copy := *attempt
	return &copy
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
