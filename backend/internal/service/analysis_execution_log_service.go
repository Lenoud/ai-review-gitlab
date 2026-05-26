package service

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

const (
	AnalysisExecutionStatusSucceeded = "succeeded"
	AnalysisExecutionStatusFailed    = "failed"
	AnalysisExecutionStatusSkipped   = "skipped"
)

type AnalysisExecutionLogSearchQuery struct {
	ProjectID uint
	PlanID    uint
	Status    string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	Size      int
}

type AnalysisTestRunInput struct {
	ProjectID         uint
	Prompt            string
	PlanName          string
	PlanID            uint
	IMEnabled         bool
	IMRobotID         uint
	HTMLReportEnabled bool
}

type AnalysisExecutionLogInput struct {
	PlanID        uint
	ProjectID     uint
	Status        string
	StartedAt     time.Time
	CompletedAt   time.Time
	DurationMs    int64
	ResultContent string
	ResultActions string
	ErrorMessage  string
	ErrorStack    string
}

type AnalysisExecutionLogPage struct {
	Items []ProjectAnalysisPlanExecutionLog `json:"items"`
	Total int64                             `json:"total"`
	Page  int                               `json:"page"`
	Size  int                               `json:"size"`
}

type AnalysisExecutionLogRepository interface {
	CreateAnalysisExecution(ctx context.Context, input AnalysisExecutionLogInput) (*ProjectAnalysisPlanExecutionLog, error)
	FindAnalysisExecutionByID(ctx context.Context, id uint) (*ProjectAnalysisPlanExecutionLog, error)
	SearchAnalysisExecution(ctx context.Context, query AnalysisExecutionLogSearchQuery) (*AnalysisExecutionLogPage, error)
	UpdateAnalysisExecutionShareToken(ctx context.Context, id uint, token string, expiresAt int64) (*ProjectAnalysisPlanExecutionLog, error)
}

type AnalysisProjectRepository interface {
	FindByID(ctx context.Context, id uint) (*Project, error)
}

type AnalysisLLMModelRepository interface {
	Default(ctx context.Context) (*LLMModel, error)
}

type AnalysisExecutionLogService struct {
	logs     AnalysisExecutionLogRepository
	projects AnalysisProjectRepository
	models   AnalysisLLMModelRepository
	llm      LLMChatClient
	now      func() time.Time
}

func NewAnalysisExecutionLogService(logs AnalysisExecutionLogRepository) *AnalysisExecutionLogService {
	return NewAnalysisExecutionLogServiceWithRunner(logs, nil, nil, nil, time.Now)
}

func NewAnalysisExecutionLogServiceWithRunner(
	logs AnalysisExecutionLogRepository,
	projects AnalysisProjectRepository,
	models AnalysisLLMModelRepository,
	llm LLMChatClient,
	now func() time.Time,
) *AnalysisExecutionLogService {
	if now == nil {
		now = time.Now
	}
	return &AnalysisExecutionLogService{
		logs:     logs,
		projects: projects,
		models:   models,
		llm:      llm,
		now:      now,
	}
}

func (s *AnalysisExecutionLogService) Get(ctx context.Context, id uint) (*ProjectAnalysisPlanExecutionLog, error) {
	if id == 0 {
		return nil, ErrInvalidReviewLogInput
	}
	return s.logs.FindAnalysisExecutionByID(ctx, id)
}

func (s *AnalysisExecutionLogService) Search(ctx context.Context, query AnalysisExecutionLogSearchQuery) (*AnalysisExecutionLogPage, error) {
	return s.logs.SearchAnalysisExecution(ctx, normalizeAnalysisExecutionLogSearchQuery(query))
}

func (s *AnalysisExecutionLogService) GenerateShareToken(ctx context.Context, id uint) (*ReviewLogShareToken, error) {
	if id == 0 {
		return nil, ErrInvalidReviewLogInput
	}
	token, err := generateReviewShareToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour).UnixMilli()
	log, err := s.logs.UpdateAnalysisExecutionShareToken(ctx, id, token, expiresAt)
	if err != nil {
		return nil, err
	}
	return &ReviewLogShareToken{ShareToken: log.ShareToken, ShareTokenExpiresAt: log.ShareTokenExpiresAt}, nil
}

func (s *AnalysisExecutionLogService) TestRun(ctx context.Context, input AnalysisTestRunInput) (*ProjectAnalysisPlanExecutionLog, error) {
	normalized, err := normalizeAnalysisTestRunInput(input)
	if err != nil {
		return nil, err
	}
	if s.projects == nil || s.models == nil || s.llm == nil {
		return nil, ErrInvalidReviewLogInput
	}
	project, err := s.projects.FindByID(ctx, normalized.ProjectID)
	if err != nil {
		return nil, err
	}
	startedAt := s.now()
	model, err := s.models.Default(ctx)
	if err != nil {
		completedAt := s.now()
		return s.recordAnalysisTestRunFailure(ctx, normalized, startedAt, completedAt, err)
	}

	result, runErr := s.llm.Chat(ctx, LLMChatInput{
		APIBaseURL: model.APIBaseURL,
		APIKey:     model.APIKey,
		ModelCode:  model.ModelCode,
		MaxTokens:  model.MaxTokens,
		Messages: []LLMChatMessage{
			{Role: "system", Content: "你是一个严谨的项目分析助手。请基于用户提示输出清晰的项目分析结论。"},
			{Role: "user", Content: buildAnalysisTestRunPrompt(project, normalized)},
		},
	})
	completedAt := s.now()
	logInput := AnalysisExecutionLogInput{
		PlanID:      normalized.PlanID,
		ProjectID:   normalized.ProjectID,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		DurationMs:  completedAt.Sub(startedAt).Milliseconds(),
	}
	if runErr != nil {
		return s.recordAnalysisTestRunFailure(ctx, normalized, startedAt, completedAt, runErr)
	}
	logInput.Status = AnalysisExecutionStatusSucceeded
	logInput.ResultContent = strings.TrimSpace(result)
	return s.logs.CreateAnalysisExecution(ctx, logInput)
}

func (s *AnalysisExecutionLogService) recordAnalysisTestRunFailure(ctx context.Context, input AnalysisTestRunInput, startedAt time.Time, completedAt time.Time, cause error) (*ProjectAnalysisPlanExecutionLog, error) {
	return s.logs.CreateAnalysisExecution(ctx, AnalysisExecutionLogInput{
		PlanID:       input.PlanID,
		ProjectID:    input.ProjectID,
		Status:       AnalysisExecutionStatusFailed,
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
		DurationMs:   completedAt.Sub(startedAt).Milliseconds(),
		ErrorMessage: cause.Error(),
		ErrorStack:   cause.Error() + "\n" + string(debug.Stack()),
	})
}

func normalizeAnalysisExecutionLogSearchQuery(query AnalysisExecutionLogSearchQuery) AnalysisExecutionLogSearchQuery {
	query.Status = strings.TrimSpace(query.Status)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 20
	}
	if query.Size > 200 {
		query.Size = 200
	}
	return query
}

func normalizeAnalysisTestRunInput(input AnalysisTestRunInput) (AnalysisTestRunInput, error) {
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.PlanName = strings.TrimSpace(input.PlanName)
	if input.ProjectID == 0 || input.Prompt == "" {
		return AnalysisTestRunInput{}, ErrInvalidReviewLogInput
	}
	return input, nil
}

func buildAnalysisTestRunPrompt(project *Project, input AnalysisTestRunInput) string {
	planName := input.PlanName
	if planName == "" {
		planName = "未命名计划"
	}
	return fmt.Sprintf("项目：%s\n计划：%s\n\n%s", strings.TrimSpace(project.Name), planName, input.Prompt)
}
