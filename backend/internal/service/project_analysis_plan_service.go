package service

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrProjectAnalysisPlanNotFound     = errors.New("project analysis plan not found")
	ErrInvalidProjectAnalysisPlanInput = errors.New("invalid project analysis plan input")
)

type ProjectAnalysisPlan struct {
	ID                uint   `json:"id"`
	ProjectID         uint   `json:"projectId"`
	Name              string `json:"name"`
	Prompt            string `json:"prompt"`
	CronExpression    string `json:"cronExpression"`
	Enabled           bool   `json:"enabled"`
	IMEnabled         bool   `json:"imEnabled"`
	IMRobotID         uint   `json:"imRobotId"`
	HTMLReportEnabled bool   `json:"htmlReportEnabled"`
	CreatedAt         int64  `json:"createdAt"`
	UpdatedAt         int64  `json:"updatedAt"`
}

type ProjectAnalysisPlanInput struct {
	ProjectID         uint
	Name              string
	Prompt            string
	CronExpression    string
	Enabled           *bool
	IMEnabled         bool
	IMRobotID         uint
	HTMLReportEnabled *bool
}

type ProjectAnalysisPlanSearchQuery struct {
	ProjectID uint
	Keyword   string
	Enabled   *bool
	Page      int
	Size      int
}

type ProjectAnalysisPlanPage struct {
	Items []ProjectAnalysisPlan `json:"items"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
}

type ProjectAnalysisPlanRepository interface {
	CreateProjectAnalysisPlan(ctx context.Context, input ProjectAnalysisPlanInput) (*ProjectAnalysisPlan, error)
	UpdateProjectAnalysisPlan(ctx context.Context, id uint, input ProjectAnalysisPlanInput) (*ProjectAnalysisPlan, error)
	FindProjectAnalysisPlanByID(ctx context.Context, id uint) (*ProjectAnalysisPlan, error)
	DeleteProjectAnalysisPlans(ctx context.Context, ids []uint) error
	SearchProjectAnalysisPlans(ctx context.Context, query ProjectAnalysisPlanSearchQuery) (*ProjectAnalysisPlanPage, error)
}

type ProjectAnalysisPlanService struct {
	plans ProjectAnalysisPlanRepository
}

func NewProjectAnalysisPlanService(plans ProjectAnalysisPlanRepository) *ProjectAnalysisPlanService {
	return &ProjectAnalysisPlanService{plans: plans}
}

func (s *ProjectAnalysisPlanService) Create(ctx context.Context, input ProjectAnalysisPlanInput) (*ProjectAnalysisPlan, error) {
	normalized, err := normalizeProjectAnalysisPlanInput(input)
	if err != nil {
		return nil, err
	}
	return s.plans.CreateProjectAnalysisPlan(ctx, normalized)
}

func (s *ProjectAnalysisPlanService) Update(ctx context.Context, id uint, input ProjectAnalysisPlanInput) (*ProjectAnalysisPlan, error) {
	if id == 0 {
		return nil, ErrInvalidProjectAnalysisPlanInput
	}
	normalized, err := normalizeProjectAnalysisPlanInput(input)
	if err != nil {
		return nil, err
	}
	return s.plans.UpdateProjectAnalysisPlan(ctx, id, normalized)
}

func (s *ProjectAnalysisPlanService) Get(ctx context.Context, id uint) (*ProjectAnalysisPlan, error) {
	if id == 0 {
		return nil, ErrInvalidProjectAnalysisPlanInput
	}
	return s.plans.FindProjectAnalysisPlanByID(ctx, id)
}

func (s *ProjectAnalysisPlanService) Delete(ctx context.Context, ids []uint) error {
	cleanIDs := make([]uint, 0, len(ids))
	seen := map[uint]struct{}{}
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		cleanIDs = append(cleanIDs, id)
	}
	if len(cleanIDs) == 0 {
		return ErrInvalidProjectAnalysisPlanInput
	}
	return s.plans.DeleteProjectAnalysisPlans(ctx, cleanIDs)
}

func (s *ProjectAnalysisPlanService) Search(ctx context.Context, query ProjectAnalysisPlanSearchQuery) (*ProjectAnalysisPlanPage, error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 20
	}
	if query.Size > 200 {
		query.Size = 200
	}
	return s.plans.SearchProjectAnalysisPlans(ctx, query)
}

func normalizeProjectAnalysisPlanInput(input ProjectAnalysisPlanInput) (ProjectAnalysisPlanInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.CronExpression = strings.TrimSpace(input.CronExpression)
	if input.ProjectID == 0 || input.Name == "" {
		return ProjectAnalysisPlanInput{}, ErrInvalidProjectAnalysisPlanInput
	}
	if input.Enabled == nil {
		enabled := true
		input.Enabled = &enabled
	}
	if input.HTMLReportEnabled == nil {
		htmlReportEnabled := true
		input.HTMLReportEnabled = &htmlReportEnabled
	}
	return input, nil
}
