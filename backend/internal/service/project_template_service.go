package service

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrProjectTemplateNotFound     = errors.New("project template not found")
	ErrInvalidProjectTemplateInput = errors.New("invalid project template input")
	ErrProjectTemplateInUse        = errors.New("project template in use")
)

type ProjectTemplate struct {
	ID                   uint     `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Extensions           []string `json:"extensions"`
	ReviewPromptTemplate string   `json:"reviewPromptTemplate"`
	CreatedAt            int64    `json:"createdAt"`
	UpdatedAt            int64    `json:"updatedAt"`
}

type ProjectTemplateInput struct {
	Name                 string
	Description          string
	Extensions           []string
	ReviewPromptTemplate string
}

type ProjectTemplateListQuery struct {
	Keyword string
}

type ProjectTemplateRepository interface {
	CreateProjectTemplate(ctx context.Context, input ProjectTemplateInput) (*ProjectTemplate, error)
	UpdateProjectTemplate(ctx context.Context, id uint, input ProjectTemplateInput) (*ProjectTemplate, error)
	FindProjectTemplateByID(ctx context.Context, id uint) (*ProjectTemplate, error)
	DeleteProjectTemplates(ctx context.Context, ids []uint) error
	ListProjectTemplates(ctx context.Context, query ProjectTemplateListQuery) ([]ProjectTemplate, error)
	CountProjectsUsingTemplates(ctx context.Context, ids []uint) (int64, error)
}

type ProjectTemplateService struct {
	templates ProjectTemplateRepository
}

func NewProjectTemplateService(templates ProjectTemplateRepository) *ProjectTemplateService {
	return &ProjectTemplateService{templates: templates}
}

func (s *ProjectTemplateService) Create(ctx context.Context, input ProjectTemplateInput) (*ProjectTemplate, error) {
	normalized, err := normalizeProjectTemplateInput(input)
	if err != nil {
		return nil, err
	}
	return s.templates.CreateProjectTemplate(ctx, normalized)
}

func (s *ProjectTemplateService) Update(ctx context.Context, id uint, input ProjectTemplateInput) (*ProjectTemplate, error) {
	if id == 0 {
		return nil, ErrInvalidProjectTemplateInput
	}
	normalized, err := normalizeProjectTemplateInput(input)
	if err != nil {
		return nil, err
	}
	return s.templates.UpdateProjectTemplate(ctx, id, normalized)
}

func (s *ProjectTemplateService) Get(ctx context.Context, id uint) (*ProjectTemplate, error) {
	if id == 0 {
		return nil, ErrInvalidProjectTemplateInput
	}
	return s.templates.FindProjectTemplateByID(ctx, id)
}

func (s *ProjectTemplateService) Delete(ctx context.Context, ids []uint) error {
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
		return ErrInvalidProjectTemplateInput
	}
	count, err := s.templates.CountProjectsUsingTemplates(ctx, cleanIDs)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrProjectTemplateInUse
	}
	return s.templates.DeleteProjectTemplates(ctx, cleanIDs)
}

func (s *ProjectTemplateService) List(ctx context.Context, query ProjectTemplateListQuery) ([]ProjectTemplate, error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	return s.templates.ListProjectTemplates(ctx, query)
}

func normalizeProjectTemplateInput(input ProjectTemplateInput) (ProjectTemplateInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.ReviewPromptTemplate = strings.TrimSpace(input.ReviewPromptTemplate)
	input.Extensions = normalizeTemplateExtensions(input.Extensions)
	if input.Name == "" {
		return ProjectTemplateInput{}, ErrInvalidProjectTemplateInput
	}
	return input, nil
}

func normalizeTemplateExtensions(values []string) []string {
	extensions := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		extensions = append(extensions, value)
	}
	return extensions
}
