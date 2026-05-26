package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

const (
	ProjectPlatformGitLab = "gitlab"

	defaultReviewPromptTemplate = `你是一位资深的软件开发工程师，专注于代码的功能正确性、安全性、稳定性以及工程最佳实践。你的任务是对员工提交的代码进行专业、克制且高价值的代码审查。

### 代码审查目标与评分权重：
1. 功能实现的正确性与健壮性（40分）：逻辑是否正确，是否能正确处理边界情况与异常场景。
2. 安全性与潜在风险（30分）：是否存在安全隐患（如 SQL 注入、XSS、越权、敏感信息泄露等）。
3. 最佳实践与可维护性（20分）：是否符合主流工程最佳实践（结构、命名、可读性、注释）。
4. 性能与资源利用（5分）：是否存在明显性能瓶颈或不必要的资源浪费。
5. 提交信息质量（5分）：commit 信息是否清晰、准确、可追溯。

### 重要规则（必须严格遵守）：
- 请仅关注并输出最重要的前三个问题（Top 3），不得多于 3 个。
- 若问题不足 3 个，则按实际数量输出。

### 输出格式（Markdown）：
请严格按照以下结构输出：

#### 一、关键问题与优化建议，如果必要，给出代码示例（仅限 Top 3）
- 按重要性从高到低排序（问题 1 最重要）。
- 每个问题需包含：问题描述、影响分析、优化建议。

#### 二、评分明细
- 按五个评分维度分别给出具体分数，并简要说明理由。`
)

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

type ReviewPrompt struct {
	ProjectID      uint   `json:"projectId,omitempty"`
	PromptTemplate string `json:"promptTemplate"`
	Customized     bool   `json:"customized"`
}

type ReviewPromptUpdateInput struct {
	ProjectID      uint
	PromptTemplate string
}

type ReviewPromptTestInput struct {
	ProjectID      uint
	PromptTemplate string
	Diffs          string
	Commits        string
}

type ReviewPromptTestResult struct {
	RenderedPrompt       string `json:"renderedPrompt"`
	CharacterCount       int    `json:"characterCount"`
	HasRequiredVariables bool   `json:"hasRequiredVariables"`
	MissingVariables     string `json:"missingVariables"`
}

type ProjectRepository interface {
	Create(ctx context.Context, input ProjectInput) (*Project, error)
	BatchCreate(ctx context.Context, inputs []ProjectInput) ([]Project, error)
	Update(ctx context.Context, id uint, input ProjectInput) (*Project, error)
	FindByID(ctx context.Context, id uint) (*Project, error)
	FindByWebURL(ctx context.Context, webURL string) (*Project, error)
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

func (s *ProjectService) GetReviewPrompt(ctx context.Context, id uint) (*ReviewPrompt, error) {
	if id == 0 {
		return nil, ErrInvalidProjectInput
	}
	project, err := s.projects.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	template := strings.TrimSpace(project.ReviewPromptTemplate)
	if template != "" {
		return &ReviewPrompt{ProjectID: project.ID, PromptTemplate: template, Customized: true}, nil
	}
	return &ReviewPrompt{ProjectID: project.ID, PromptTemplate: defaultReviewPromptTemplate, Customized: false}, nil
}

func (s *ProjectService) GetDefaultReviewPrompt(ctx context.Context) *ReviewPrompt {
	return &ReviewPrompt{PromptTemplate: defaultReviewPromptTemplate, Customized: false}
}

func (s *ProjectService) UpdateReviewPrompt(ctx context.Context, input ReviewPromptUpdateInput) (*ReviewPrompt, error) {
	input.PromptTemplate = strings.TrimSpace(input.PromptTemplate)
	if input.ProjectID == 0 || input.PromptTemplate == "" {
		return nil, ErrInvalidProjectInput
	}
	project, err := s.projects.FindByID(ctx, input.ProjectID)
	if err != nil {
		return nil, err
	}
	project.ReviewPromptTemplate = input.PromptTemplate
	aiReviewEnabled := project.AIReviewEnabled
	updated, err := s.projects.Update(ctx, input.ProjectID, ProjectInput{
		Name:                     project.Name,
		Description:              project.Description,
		WebURL:                   project.WebURL,
		Platform:                 project.Platform,
		AccessToken:              project.AccessToken,
		IMEnabled:                project.IMEnabled,
		IMRobotID:                project.IMRobotID,
		IMAtMemberEnabled:        project.IMAtMemberEnabled,
		IMAtMemberScoreThreshold: project.IMAtMemberScoreThreshold,
		AIReviewEnabled:          &aiReviewEnabled,
		TemplateID:               project.TemplateID,
		Extensions:               project.Extensions,
		ReviewEventTypes:         project.ReviewEventTypes,
		ReviewPromptTemplate:     input.PromptTemplate,
		HTMLReportEnabled:        project.HTMLReportEnabled,
		DeepReviewEnabled:        project.DeepReviewEnabled,
	})
	if err != nil {
		return nil, err
	}
	return &ReviewPrompt{ProjectID: updated.ID, PromptTemplate: updated.ReviewPromptTemplate, Customized: true}, nil
}

func (s *ProjectService) DeleteReviewPrompt(ctx context.Context, id uint) error {
	if id == 0 {
		return ErrInvalidProjectInput
	}
	project, err := s.projects.FindByID(ctx, id)
	if err != nil {
		return err
	}
	aiReviewEnabled := project.AIReviewEnabled
	_, err = s.projects.Update(ctx, id, ProjectInput{
		Name:                     project.Name,
		Description:              project.Description,
		WebURL:                   project.WebURL,
		Platform:                 project.Platform,
		AccessToken:              project.AccessToken,
		IMEnabled:                project.IMEnabled,
		IMRobotID:                project.IMRobotID,
		IMAtMemberEnabled:        project.IMAtMemberEnabled,
		IMAtMemberScoreThreshold: project.IMAtMemberScoreThreshold,
		AIReviewEnabled:          &aiReviewEnabled,
		TemplateID:               project.TemplateID,
		Extensions:               project.Extensions,
		ReviewEventTypes:         project.ReviewEventTypes,
		ReviewPromptTemplate:     "",
		HTMLReportEnabled:        project.HTMLReportEnabled,
		DeepReviewEnabled:        project.DeepReviewEnabled,
	})
	return err
}

func (s *ProjectService) TestReviewPrompt(ctx context.Context, input ReviewPromptTestInput) (*ReviewPromptTestResult, error) {
	input.PromptTemplate = strings.TrimSpace(input.PromptTemplate)
	if input.PromptTemplate == "" {
		return nil, ErrInvalidProjectInput
	}
	projectName := ""
	if input.ProjectID > 0 {
		project, err := s.projects.FindByID(ctx, input.ProjectID)
		if err == nil {
			projectName = project.Name
		} else if !errors.Is(err, ErrProjectNotFound) {
			return nil, err
		}
	}
	diffs := input.Diffs
	if strings.TrimSpace(diffs) == "" {
		diffs = "[示例代码变更]"
	}
	commits := input.Commits
	if strings.TrimSpace(commits) == "" {
		commits = "[示例提交历史]"
	}
	rendered := AssembleReviewPrompt(ReviewPromptAssembleInput{
		Template:    input.PromptTemplate,
		Diffs:       diffs,
		Commits:     commits,
		ProjectName: projectName,
	})
	return &ReviewPromptTestResult{
		RenderedPrompt:       rendered,
		CharacterCount:       len(rendered),
		HasRequiredVariables: true,
		MissingVariables:     "",
	}, nil
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
