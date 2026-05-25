package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrInvalidReviewWorkerInput = errors.New("invalid review worker input")

type ReviewWorkerTaskRunner interface {
	ClaimNext(ctx context.Context, workerID string) (*ReviewTask, error)
	StartAttempt(ctx context.Context, taskID uint) (*ReviewTaskAttempt, error)
	MarkSucceeded(ctx context.Context, taskID uint) error
	MarkFailed(ctx context.Context, taskID uint, message string) error
}

type ReviewWorkerProjectRepository interface {
	FindByID(ctx context.Context, id uint) (*Project, error)
}

type ReviewWorkerLLMModelRepository interface {
	Default(ctx context.Context) (*LLMModel, error)
}

type LLMChatClient interface {
	Chat(ctx context.Context, input LLMChatInput) (string, error)
}

type LLMChatInput struct {
	APIBaseURL string
	APIKey     string
	ModelCode  string
	MaxTokens  int
	Messages   []LLMChatMessage
}

type LLMChatMessage struct {
	Role    string
	Content string
}

type ReviewWorkerResult struct {
	Processed  bool   `json:"processed"`
	TaskID     uint   `json:"taskId,omitempty"`
	ReviewText string `json:"reviewText,omitempty"`
}

type ReviewWorkerService struct {
	tasks    ReviewWorkerTaskRunner
	projects ReviewWorkerProjectRepository
	models   ReviewWorkerLLMModelRepository
	gitlab   GitLabClient
	llm      LLMChatClient
}

func NewReviewWorkerService(
	tasks ReviewWorkerTaskRunner,
	projects ReviewWorkerProjectRepository,
	models ReviewWorkerLLMModelRepository,
	gitlab GitLabClient,
	llm LLMChatClient,
) *ReviewWorkerService {
	return &ReviewWorkerService{
		tasks:    tasks,
		projects: projects,
		models:   models,
		gitlab:   gitlab,
		llm:      llm,
	}
}

func (s *ReviewWorkerService) ProcessNext(ctx context.Context, workerID string) (*ReviewWorkerResult, error) {
	task, err := s.tasks.ClaimNext(ctx, workerID)
	if err != nil {
		if errors.Is(err, ErrReviewTaskNotFound) {
			return &ReviewWorkerResult{Processed: false}, nil
		}
		return nil, err
	}
	result := &ReviewWorkerResult{Processed: true, TaskID: task.ID}

	if _, err := s.tasks.StartAttempt(ctx, task.ID); err != nil {
		return result, err
	}

	reviewText, err := s.processClaimedTask(ctx, task)
	if err != nil {
		if markErr := s.tasks.MarkFailed(ctx, task.ID, err.Error()); markErr != nil {
			return result, markErr
		}
		return result, err
	}
	if err := s.tasks.MarkSucceeded(ctx, task.ID); err != nil {
		return result, err
	}
	result.ReviewText = reviewText
	return result, nil
}

func (s *ReviewWorkerService) processClaimedTask(ctx context.Context, task *ReviewTask) (string, error) {
	project, err := s.projects.FindByID(ctx, task.ProjectID)
	if err != nil {
		return "", err
	}
	model, err := s.models.Default(ctx)
	if err != nil {
		return "", err
	}
	diff, err := s.fetchTaskDiff(ctx, task, project)
	if err != nil {
		return "", err
	}
	return s.llm.Chat(ctx, LLMChatInput{
		APIBaseURL: model.APIBaseURL,
		APIKey:     model.APIKey,
		ModelCode:  model.ModelCode,
		MaxTokens:  model.MaxTokens,
		Messages: []LLMChatMessage{
			{Role: "system", Content: "你是一个严谨的代码审查助手。请基于 diff 给出问题、风险和建议。"},
			{Role: "user", Content: buildReviewPrompt(project, task, diff)},
		},
	})
}

func (s *ReviewWorkerService) fetchTaskDiff(ctx context.Context, task *ReviewTask, project *Project) ([]GitLabDiff, error) {
	payload, err := parseReviewWorkerPayload(task.EventType, task.PayloadJSON)
	if err != nil {
		return nil, err
	}
	baseURL, err := gitLabBaseURL(project.WebURL)
	if err != nil {
		return nil, err
	}
	client := s.gitlab.WithAuth(baseURL, project.AccessToken)
	switch task.EventType {
	case ReviewTaskEventPush:
		return client.GetCommitDiff(ctx, payload.ProjectID, payload.AfterSHA)
	case ReviewTaskEventMergeRequest:
		return client.GetMergeRequestChanges(ctx, payload.ProjectID, payload.MRIID)
	default:
		return nil, ErrInvalidReviewWorkerInput
	}
}

type reviewWorkerPayload struct {
	ProjectID int
	AfterSHA  string
	MRIID     int
}

func parseReviewWorkerPayload(eventType string, payload []byte) (*reviewWorkerPayload, error) {
	switch eventType {
	case ReviewTaskEventPush:
		var body struct {
			Project struct {
				ID int `json:"id"`
			} `json:"project"`
			After string `json:"after"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, ErrInvalidReviewWorkerInput
		}
		after := strings.TrimSpace(body.After)
		if body.Project.ID == 0 || after == "" {
			return nil, ErrInvalidReviewWorkerInput
		}
		return &reviewWorkerPayload{ProjectID: body.Project.ID, AfterSHA: after}, nil
	case ReviewTaskEventMergeRequest:
		var body struct {
			Project struct {
				ID int `json:"id"`
			} `json:"project"`
			ObjectAttributes struct {
				IID int `json:"iid"`
			} `json:"object_attributes"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, ErrInvalidReviewWorkerInput
		}
		if body.Project.ID == 0 || body.ObjectAttributes.IID == 0 {
			return nil, ErrInvalidReviewWorkerInput
		}
		return &reviewWorkerPayload{ProjectID: body.Project.ID, MRIID: body.ObjectAttributes.IID}, nil
	default:
		return nil, ErrInvalidReviewWorkerInput
	}
}

func gitLabBaseURL(webURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(webURL))
	if err != nil {
		return "", ErrInvalidReviewWorkerInput
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidReviewWorkerInput
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}

func buildReviewPrompt(project *Project, task *ReviewTask, diff []GitLabDiff) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("项目: %s\n", project.Name))
	builder.WriteString(fmt.Sprintf("事件: %s\n", task.EventType))
	if strings.TrimSpace(project.ReviewPromptTemplate) != "" {
		builder.WriteString("项目审查要求:\n")
		builder.WriteString(project.ReviewPromptTemplate)
		builder.WriteString("\n")
	}
	builder.WriteString("代码 diff:\n")
	for _, item := range diff {
		path := item.NewPath
		if path == "" {
			path = item.OldPath
		}
		builder.WriteString(fmt.Sprintf("\n--- %s ---\n", path))
		builder.WriteString(item.Diff)
		builder.WriteString("\n")
	}
	return builder.String()
}
