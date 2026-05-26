package service

import (
	"context"
	"strings"
	"time"
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

type AnalysisExecutionLogPage struct {
	Items []ProjectAnalysisPlanExecutionLog `json:"items"`
	Total int64                             `json:"total"`
	Page  int                               `json:"page"`
	Size  int                               `json:"size"`
}

type AnalysisExecutionLogRepository interface {
	FindAnalysisExecutionByID(ctx context.Context, id uint) (*ProjectAnalysisPlanExecutionLog, error)
	SearchAnalysisExecution(ctx context.Context, query AnalysisExecutionLogSearchQuery) (*AnalysisExecutionLogPage, error)
	UpdateAnalysisExecutionShareToken(ctx context.Context, id uint, token string, expiresAt int64) (*ProjectAnalysisPlanExecutionLog, error)
}

type AnalysisExecutionLogService struct {
	logs AnalysisExecutionLogRepository
}

func NewAnalysisExecutionLogService(logs AnalysisExecutionLogRepository) *AnalysisExecutionLogService {
	return &AnalysisExecutionLogService{logs: logs}
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
