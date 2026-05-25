package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
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

type ReviewWorkerLogWriter interface {
	CreatePush(ctx context.Context, input PushReviewLogInput) (*PushReviewLog, error)
	CreateMergeRequest(ctx context.Context, input MergeRequestReviewLogInput) (*MergeRequestReviewLog, error)
}

type ReviewWorkerTraceWriter interface {
	Create(ctx context.Context, input AIReviewTraceInput) (*AIReviewTrace, error)
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
	logs     ReviewWorkerLogWriter
	traces   ReviewWorkerTraceWriter
}

func NewReviewWorkerService(
	tasks ReviewWorkerTaskRunner,
	projects ReviewWorkerProjectRepository,
	models ReviewWorkerLLMModelRepository,
	gitlab GitLabClient,
	llm LLMChatClient,
	logs ReviewWorkerLogWriter,
	traces ReviewWorkerTraceWriter,
) *ReviewWorkerService {
	return &ReviewWorkerService{
		tasks:    tasks,
		projects: projects,
		models:   models,
		gitlab:   gitlab,
		llm:      llm,
		logs:     logs,
		traces:   traces,
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
	payload, diff, err := s.fetchTaskDiff(ctx, task, project)
	if err != nil {
		return "", err
	}
	messages := []LLMChatMessage{
		{Role: "system", Content: "你是一个严谨的代码审查助手。请基于 diff 给出问题、风险和建议。"},
		{Role: "user", Content: buildReviewPrompt(project, task, diff)},
	}
	reviewText, err := s.llm.Chat(ctx, LLMChatInput{
		APIBaseURL: model.APIBaseURL,
		APIKey:     model.APIKey,
		ModelCode:  model.ModelCode,
		MaxTokens:  model.MaxTokens,
		Messages:   messages,
	})
	if err != nil {
		return "", err
	}
	if s.logs != nil {
		eventType, eventID, err := s.createReviewLog(ctx, task, project, payload, diff, reviewText)
		if err != nil {
			return "", err
		}
		if s.traces != nil {
			if err := s.createAIReviewTrace(ctx, eventType, eventID, messages, reviewText, model); err != nil {
				return "", err
			}
		}
	}
	return reviewText, nil
}

func (s *ReviewWorkerService) fetchTaskDiff(ctx context.Context, task *ReviewTask, project *Project) (*reviewWorkerPayload, []GitLabDiff, error) {
	payload, err := parseReviewWorkerPayload(task.EventType, task.PayloadJSON)
	if err != nil {
		return nil, nil, err
	}
	baseURL, err := gitLabBaseURL(project.WebURL)
	if err != nil {
		return nil, nil, err
	}
	client := s.gitlab.WithAuth(baseURL, project.AccessToken)
	switch task.EventType {
	case ReviewTaskEventPush:
		diff, err := client.GetCommitDiff(ctx, payload.ProjectID, payload.AfterSHA)
		return payload, diff, err
	case ReviewTaskEventMergeRequest:
		diff, err := client.GetMergeRequestChanges(ctx, payload.ProjectID, payload.MRIID)
		return payload, diff, err
	default:
		return nil, nil, ErrInvalidReviewWorkerInput
	}
}

type reviewWorkerPayload struct {
	ProjectID         int
	AfterSHA          string
	MRIID             int
	AuthorIdentity    string
	AuthorDisplayName string
	Branch            string
	CommitMessages    string
	Commits           []ReviewCommit
	LastCommitURL     string
	SourceBranch      string
	TargetBranch      string
	LastCommitID      string
	URL               string
}

func parseReviewWorkerPayload(eventType string, payload []byte) (*reviewWorkerPayload, error) {
	switch eventType {
	case ReviewTaskEventPush:
		var body struct {
			Project struct {
				ID     int    `json:"id"`
				WebURL string `json:"web_url"`
			} `json:"project"`
			UserUsername string `json:"user_username"`
			UserName     string `json:"user_name"`
			UserEmail    string `json:"user_email"`
			Ref          string `json:"ref"`
			After        string `json:"after"`
			Commits      []struct {
				ID        string `json:"id"`
				Message   string `json:"message"`
				URL       string `json:"url"`
				Timestamp string `json:"timestamp"`
				Author    struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"author"`
			} `json:"commits"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, ErrInvalidReviewWorkerInput
		}
		after := strings.TrimSpace(body.After)
		if body.Project.ID == 0 || after == "" {
			return nil, ErrInvalidReviewWorkerInput
		}
		commits := make([]ReviewCommit, 0, len(body.Commits))
		for _, commit := range body.Commits {
			commits = append(commits, ReviewCommit{
				Author:    firstNonBlank(commit.Author.Name, commit.Author.Email),
				Message:   strings.TrimSpace(commit.Message),
				URL:       strings.TrimSpace(commit.URL),
				Timestamp: strings.TrimSpace(commit.Timestamp),
			})
		}
		return &reviewWorkerPayload{
			ProjectID:         body.Project.ID,
			AfterSHA:          after,
			AuthorIdentity:    firstNonBlank(body.UserUsername, body.UserEmail, body.UserName),
			AuthorDisplayName: firstNonBlank(body.UserName, body.UserUsername),
			Branch:            refToBranch(body.Ref),
			CommitMessages:    buildCommitMessages(commits),
			Commits:           commits,
			LastCommitURL:     buildLastCommitURL(commits, body.Project.WebURL),
		}, nil
	case ReviewTaskEventMergeRequest:
		var body struct {
			Project struct {
				ID int `json:"id"`
			} `json:"project"`
			User struct {
				Username string `json:"username"`
				Name     string `json:"name"`
			} `json:"user"`
			UserUsername     string `json:"user_username"`
			UserName         string `json:"user_name"`
			ObjectAttributes struct {
				IID          int    `json:"iid"`
				SourceBranch string `json:"source_branch"`
				TargetBranch string `json:"target_branch"`
				URL          string `json:"url"`
				LastCommit   struct {
					ID      string `json:"id"`
					Message string `json:"message"`
				} `json:"last_commit"`
			} `json:"object_attributes"`
		}
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, ErrInvalidReviewWorkerInput
		}
		if body.Project.ID == 0 || body.ObjectAttributes.IID == 0 {
			return nil, ErrInvalidReviewWorkerInput
		}
		return &reviewWorkerPayload{
			ProjectID:         body.Project.ID,
			MRIID:             body.ObjectAttributes.IID,
			AuthorIdentity:    firstNonBlank(body.User.Username, body.UserUsername, body.User.Name, body.UserName),
			AuthorDisplayName: firstNonBlank(body.User.Name, body.UserName, body.User.Username, body.UserUsername),
			SourceBranch:      strings.TrimSpace(body.ObjectAttributes.SourceBranch),
			TargetBranch:      strings.TrimSpace(body.ObjectAttributes.TargetBranch),
			CommitMessages:    strings.TrimSpace(body.ObjectAttributes.LastCommit.Message),
			LastCommitID:      strings.TrimSpace(body.ObjectAttributes.LastCommit.ID),
			URL:               strings.TrimSpace(body.ObjectAttributes.URL),
		}, nil
	default:
		return nil, ErrInvalidReviewWorkerInput
	}
}

func (s *ReviewWorkerService) createReviewLog(ctx context.Context, task *ReviewTask, project *Project, payload *reviewWorkerPayload, diff []GitLabDiff, reviewText string) (string, uint, error) {
	additions, deletions := countDiffStats(diff)
	score := parseReviewScore(reviewText)
	switch task.EventType {
	case ReviewTaskEventPush:
		log, err := s.logs.CreatePush(ctx, PushReviewLogInput{
			ProjectID:         project.ID,
			ProjectName:       project.Name,
			Author:            payload.AuthorIdentity,
			AuthorIdentity:    payload.AuthorIdentity,
			AuthorDisplayName: payload.AuthorDisplayName,
			Branch:            payload.Branch,
			CommitMessages:    payload.CommitMessages,
			Commits:           payload.Commits,
			Score:             score,
			Additions:         additions,
			Deletions:         deletions,
			LastCommitURL:     payload.LastCommitURL,
			ReviewResult:      reviewText,
		})
		if err != nil {
			return "", 0, err
		}
		return ReviewTaskEventPush, log.ID, nil
	case ReviewTaskEventMergeRequest:
		log, err := s.logs.CreateMergeRequest(ctx, MergeRequestReviewLogInput{
			ProjectID:         project.ID,
			ProjectName:       project.Name,
			Author:            payload.AuthorIdentity,
			AuthorIdentity:    payload.AuthorIdentity,
			AuthorDisplayName: payload.AuthorDisplayName,
			SourceBranch:      payload.SourceBranch,
			TargetBranch:      payload.TargetBranch,
			CommitMessages:    payload.CommitMessages,
			Score:             score,
			Additions:         additions,
			Deletions:         deletions,
			LastCommitID:      payload.LastCommitID,
			URL:               payload.URL,
			ReviewResult:      reviewText,
		})
		if err != nil {
			return "", 0, err
		}
		return ReviewTaskEventMergeRequest, log.ID, nil
	default:
		return "", 0, ErrInvalidReviewWorkerInput
	}
}

func (s *ReviewWorkerService) createAIReviewTrace(ctx context.Context, eventType string, eventID uint, messages []LLMChatMessage, reviewText string, model *LLMModel) error {
	if eventID == 0 {
		return ErrInvalidReviewWorkerInput
	}
	_, err := s.traces.Create(ctx, AIReviewTraceInput{
		ReviewEventType: eventType,
		ReviewEventID:   eventID,
		Prompt:          renderLLMPrompt(messages),
		Response:        reviewText,
		Provider:        model.Provider,
		ModelCode:       model.ModelCode,
	})
	return err
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

func renderLLMPrompt(messages []LLMChatMessage) string {
	var builder strings.Builder
	for i, message := range messages {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(strings.TrimSpace(message.Role))
		builder.WriteString(":\n")
		builder.WriteString(message.Content)
	}
	return builder.String()
}

func refToBranch(ref string) string {
	ref = strings.TrimSpace(ref)
	return strings.TrimPrefix(ref, "refs/heads/")
}

func buildCommitMessages(commits []ReviewCommit) string {
	var builder strings.Builder
	for _, commit := range commits {
		message := strings.TrimRight(strings.TrimSpace(commit.Message), "\n")
		author := strings.TrimSpace(commit.Author)
		if message == "" && author == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("%s (by %s);", message, author))
	}
	return builder.String()
}

func buildLastCommitURL(commits []ReviewCommit, fallback string) string {
	for i := len(commits) - 1; i >= 0; i-- {
		if strings.TrimSpace(commits[i].URL) != "" {
			return strings.TrimSpace(commits[i].URL)
		}
	}
	return strings.TrimSpace(fallback)
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func countDiffStats(diff []GitLabDiff) (int, int) {
	additions := 0
	deletions := 0
	for _, item := range diff {
		for _, line := range strings.Split(item.Diff, "\n") {
			if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
				continue
			}
			if strings.HasPrefix(line, "+") {
				additions++
			}
			if strings.HasPrefix(line, "-") {
				deletions++
			}
		}
	}
	return additions, deletions
}

func parseReviewScore(reviewText string) int {
	matches := regexp.MustCompile(`总分[:：]\s*(\d+)分?`).FindStringSubmatch(reviewText)
	if len(matches) != 2 {
		return 0
	}
	var score int
	if _, err := fmt.Sscanf(matches[1], "%d", &score); err != nil {
		return 0
	}
	if score < 0 || score > 100 {
		return 0
	}
	return score
}
