package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

const ProjectPlatformGitLab = "gitlab"

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrProjectWebURLExists = errors.New("project web url exists")
	ErrInvalidProjectInput = errors.New("invalid project input")
)

type Project struct {
	ID                       uint     `json:"id"`
	Name                     string   `json:"name"`
	Description              string   `json:"description"`
	WebURL                   string   `json:"webUrl"`
	Platform                 string   `json:"platform"`
	AccessToken              string   `json:"accessToken,omitempty"`
	IMEnabled                bool     `json:"imEnabled"`
	IMRobotID                uint     `json:"imRobotId"`
	IMAtMemberEnabled        bool     `json:"imAtMemberEnabled"`
	IMAtMemberScoreThreshold int      `json:"imAtMemberScoreThreshold"`
	AIReviewEnabled          bool     `json:"aiReviewEnabled"`
	TemplateID               uint     `json:"templateId"`
	Extensions               []string `json:"extensions"`
	ReviewEventTypes         []string `json:"reviewEventTypes"`
	ReviewPromptTemplate     string   `json:"reviewPromptTemplate"`
	HTMLReportEnabled        bool     `json:"htmlReportEnabled"`
	DeepReviewEnabled        bool     `json:"deepReviewEnabled"`
}

type ProjectInput struct {
	Name                     string
	Description              string
	WebURL                   string
	Platform                 string
	AccessToken              string
	IMEnabled                bool
	IMRobotID                uint
	IMAtMemberEnabled        bool
	IMAtMemberScoreThreshold int
	AIReviewEnabled          *bool
	TemplateID               uint
	Extensions               []string
	ReviewEventTypes         []string
	ReviewPromptTemplate     string
	HTMLReportEnabled        bool
	DeepReviewEnabled        bool
}

type ProjectSearchQuery struct {
	Keyword  string
	Platform string
	Page     int
	Size     int
}

type ProjectPage struct {
	Items []Project `json:"items"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}

type ProjectRepository interface {
	Create(ctx context.Context, input ProjectInput) (*Project, error)
	BatchCreate(ctx context.Context, inputs []ProjectInput) ([]Project, error)
	Update(ctx context.Context, id uint, input ProjectInput) (*Project, error)
	FindByID(ctx context.Context, id uint) (*Project, error)
	Delete(ctx context.Context, ids []uint) error
	Search(ctx context.Context, query ProjectSearchQuery) (*ProjectPage, error)
	ExistsByWebURL(ctx context.Context, webURL string, excludeID uint) (bool, error)
}

type ProjectService struct {
	projects ProjectRepository
}

func NewProjectService(projects ProjectRepository) *ProjectService {
	return &ProjectService{projects: projects}
}

func (s *ProjectService) Create(ctx context.Context, input ProjectInput) (*Project, error) {
	normalized, err := normalizeProjectInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.projects.ExistsByWebURL(ctx, normalized.WebURL, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProjectWebURLExists
	}
	return s.projects.Create(ctx, normalized)
}

func (s *ProjectService) BatchCreate(ctx context.Context, inputs []ProjectInput) ([]Project, error) {
	if len(inputs) == 0 {
		return nil, ErrInvalidProjectInput
	}
	normalized := make([]ProjectInput, 0, len(inputs))
	seen := map[string]struct{}{}
	for _, input := range inputs {
		item, err := normalizeProjectInput(input)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[item.WebURL]; ok {
			return nil, ErrProjectWebURLExists
		}
		exists, err := s.projects.ExistsByWebURL(ctx, item.WebURL, 0)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrProjectWebURLExists
		}
		seen[item.WebURL] = struct{}{}
		normalized = append(normalized, item)
	}
	return s.projects.BatchCreate(ctx, normalized)
}

func (s *ProjectService) Update(ctx context.Context, id uint, input ProjectInput) (*Project, error) {
	if id == 0 {
		return nil, ErrInvalidProjectInput
	}
	if _, err := s.projects.FindByID(ctx, id); err != nil {
		return nil, err
	}
	normalized, err := normalizeProjectInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.projects.ExistsByWebURL(ctx, normalized.WebURL, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProjectWebURLExists
	}
	return s.projects.Update(ctx, id, normalized)
}

func (s *ProjectService) Get(ctx context.Context, id uint) (*Project, error) {
	if id == 0 {
		return nil, ErrInvalidProjectInput
	}
	return s.projects.FindByID(ctx, id)
}

func (s *ProjectService) Delete(ctx context.Context, ids []uint) error {
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
		return ErrInvalidProjectInput
	}
	return s.projects.Delete(ctx, cleanIDs)
}

func (s *ProjectService) Search(ctx context.Context, query ProjectSearchQuery) (*ProjectPage, error) {
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Platform = strings.TrimSpace(query.Platform)
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 {
		query.Size = 20
	}
	if query.Size > 200 {
		query.Size = 200
	}
	return s.projects.Search(ctx, query)
}

func (s *ProjectService) WebURLExists(ctx context.Context, webURL string, excludeID uint) (bool, error) {
	webURL = normalizeWebURL(webURL)
	if webURL == "" {
		return false, ErrInvalidProjectInput
	}
	return s.projects.ExistsByWebURL(ctx, webURL, excludeID)
}

func normalizeProjectInput(input ProjectInput) (ProjectInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.WebURL = normalizeWebURL(input.WebURL)
	input.Platform = strings.TrimSpace(input.Platform)
	input.Description = strings.TrimSpace(input.Description)
	input.AccessToken = strings.TrimSpace(input.AccessToken)
	input.ReviewPromptTemplate = strings.TrimSpace(input.ReviewPromptTemplate)
	input.Extensions = cleanStringSlice(input.Extensions)
	input.ReviewEventTypes = cleanStringSlice(input.ReviewEventTypes)
	if input.Name == "" || input.WebURL == "" {
		return ProjectInput{}, ErrInvalidProjectInput
	}
	if input.Platform == "" {
		input.Platform = ProjectPlatformGitLab
	}
	if _, err := url.ParseRequestURI(input.WebURL); err != nil {
		return ProjectInput{}, ErrInvalidProjectInput
	}
	if input.AIReviewEnabled == nil {
		enabled := true
		input.AIReviewEnabled = &enabled
	}
	return input, nil
}

func cleanStringSlice(values []string) []string {
	cleaned := make([]string, 0, len(values))
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
		cleaned = append(cleaned, value)
	}
	return cleaned
}

func normalizeWebURL(webURL string) string {
	return strings.TrimRight(strings.TrimSpace(webURL), "/")
}

func projectFromInput(input ProjectInput) *Project {
	aiReviewEnabled := true
	if input.AIReviewEnabled != nil {
		aiReviewEnabled = *input.AIReviewEnabled
	}
	return &Project{
		Name:                     input.Name,
		Description:              input.Description,
		WebURL:                   input.WebURL,
		Platform:                 input.Platform,
		AccessToken:              input.AccessToken,
		IMEnabled:                input.IMEnabled,
		IMRobotID:                input.IMRobotID,
		IMAtMemberEnabled:        input.IMAtMemberEnabled,
		IMAtMemberScoreThreshold: input.IMAtMemberScoreThreshold,
		AIReviewEnabled:          aiReviewEnabled,
		TemplateID:               input.TemplateID,
		Extensions:               append([]string(nil), input.Extensions...),
		ReviewEventTypes:         append([]string(nil), input.ReviewEventTypes...),
		ReviewPromptTemplate:     input.ReviewPromptTemplate,
		HTMLReportEnabled:        input.HTMLReportEnabled,
		DeepReviewEnabled:        input.DeepReviewEnabled,
	}
}

func cloneProject(project *Project) *Project {
	if project == nil {
		return nil
	}
	copy := *project
	copy.Extensions = append([]string(nil), project.Extensions...)
	copy.ReviewEventTypes = append([]string(nil), project.ReviewEventTypes...)
	return &copy
}

func containsFold(s string, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
