package service

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrProjectTemplateReviewRuleNotFound         = errors.New("project template review rule not found")
	ErrInvalidProjectTemplateReviewRuleInput     = errors.New("invalid project template review rule input")
	ErrProjectTemplateReviewRuleTemplateMismatch = errors.New("project template review rule template mismatch")
)

type ProjectTemplateReviewRule struct {
	ID           uint     `json:"id"`
	TemplateID   uint     `json:"templateId"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	GlobPatterns []string `json:"globPatterns"`
	Content      string   `json:"content"`
	Priority     int      `json:"priority"`
	Enabled      bool     `json:"enabled"`
	CreatedAt    int64    `json:"createdAt"`
	UpdatedAt    int64    `json:"updatedAt"`
}

type ProjectTemplateReviewRuleInput struct {
	TemplateID   uint
	Name         string
	Description  string
	GlobPatterns []string
	Content      string
	Priority     int
	Enabled      bool
	EnabledIsSet bool
}

type ProjectTemplateReviewRuleRepository interface {
	CreateProjectTemplateReviewRule(ctx context.Context, input ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error)
	UpdateProjectTemplateReviewRule(ctx context.Context, id uint, input ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error)
	FindProjectTemplateReviewRuleByID(ctx context.Context, id uint) (*ProjectTemplateReviewRule, error)
	DeleteProjectTemplateReviewRule(ctx context.Context, id uint) error
	ListProjectTemplateReviewRulesByTemplateID(ctx context.Context, templateID uint) ([]ProjectTemplateReviewRule, error)
	ProjectTemplateExists(ctx context.Context, templateID uint) (bool, error)
}

type ProjectTemplateReviewRuleService struct {
	rules ProjectTemplateReviewRuleRepository
}

func NewProjectTemplateReviewRuleService(rules ProjectTemplateReviewRuleRepository) *ProjectTemplateReviewRuleService {
	return &ProjectTemplateReviewRuleService{rules: rules}
}

func (s *ProjectTemplateReviewRuleService) Create(ctx context.Context, input ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error) {
	normalized, err := normalizeProjectTemplateReviewRuleInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.ensureTemplateExists(ctx, normalized.TemplateID); err != nil {
		return nil, err
	}
	return s.rules.CreateProjectTemplateReviewRule(ctx, normalized)
}

func (s *ProjectTemplateReviewRuleService) Update(ctx context.Context, id uint, input ProjectTemplateReviewRuleInput) (*ProjectTemplateReviewRule, error) {
	if id == 0 {
		return nil, ErrInvalidProjectTemplateReviewRuleInput
	}
	normalized, err := normalizeProjectTemplateReviewRuleInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.ensureTemplateExists(ctx, normalized.TemplateID); err != nil {
		return nil, err
	}
	existing, err := s.rules.FindProjectTemplateReviewRuleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.TemplateID != normalized.TemplateID {
		return nil, ErrProjectTemplateReviewRuleTemplateMismatch
	}
	return s.rules.UpdateProjectTemplateReviewRule(ctx, id, normalized)
}

func (s *ProjectTemplateReviewRuleService) Get(ctx context.Context, id uint) (*ProjectTemplateReviewRule, error) {
	if id == 0 {
		return nil, ErrInvalidProjectTemplateReviewRuleInput
	}
	return s.rules.FindProjectTemplateReviewRuleByID(ctx, id)
}

func (s *ProjectTemplateReviewRuleService) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrInvalidProjectTemplateReviewRuleInput
	}
	return s.rules.DeleteProjectTemplateReviewRule(ctx, id)
}

func (s *ProjectTemplateReviewRuleService) ListByTemplateID(ctx context.Context, templateID uint) ([]ProjectTemplateReviewRule, error) {
	if templateID == 0 {
		return nil, ErrInvalidProjectTemplateReviewRuleInput
	}
	return s.rules.ListProjectTemplateReviewRulesByTemplateID(ctx, templateID)
}

func (s *ProjectTemplateReviewRuleService) ensureTemplateExists(ctx context.Context, templateID uint) error {
	if templateID == 0 {
		return ErrInvalidProjectTemplateReviewRuleInput
	}
	exists, err := s.rules.ProjectTemplateExists(ctx, templateID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrProjectTemplateNotFound
	}
	return nil
}

func normalizeProjectTemplateReviewRuleInput(input ProjectTemplateReviewRuleInput) (ProjectTemplateReviewRuleInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Content = strings.TrimSpace(input.Content)
	input.GlobPatterns = normalizeTemplateExtensions(input.GlobPatterns)
	if !input.EnabledIsSet {
		input.Enabled = true
	}
	if input.TemplateID == 0 || input.Name == "" || len(input.GlobPatterns) == 0 || input.Content == "" {
		return ProjectTemplateReviewRuleInput{}, ErrInvalidProjectTemplateReviewRuleInput
	}
	return input, nil
}
