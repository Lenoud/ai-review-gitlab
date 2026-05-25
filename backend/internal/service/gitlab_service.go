package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidGitLabInput = errors.New("invalid gitlab input")

type GitLabSearchInput struct {
	APIBaseURL  string
	AccessToken string
	Keyword     string
	Page        int
	Size        int
}

type GitLabSearchOptions struct {
	Keyword string
	Page    int
	PerPage int
}

type GitLabProject struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"pathWithNamespace"`
	WebURL            string `json:"webUrl"`
	DefaultBranch     string `json:"defaultBranch"`
}

type GitLabGroup struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullPath string `json:"fullPath"`
	WebURL   string `json:"webUrl"`
}

type GitLabDiff struct {
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
	Diff    string `json:"diff"`
}

type GitLabClient interface {
	WithAuth(baseURL string, token string) GitLabClient
	SearchProjects(ctx context.Context, opts GitLabSearchOptions) ([]GitLabProject, error)
	SearchGroups(ctx context.Context, opts GitLabSearchOptions) ([]GitLabGroup, error)
	GetMergeRequestChanges(ctx context.Context, projectID int, mergeRequestIID int) ([]GitLabDiff, error)
	GetCommitDiff(ctx context.Context, projectID int, sha string) ([]GitLabDiff, error)
}

type GitLabService struct {
	client GitLabClient
}

func NewGitLabService(client GitLabClient) *GitLabService {
	return &GitLabService{client: client}
}

func (s *GitLabService) SearchProjects(ctx context.Context, input GitLabSearchInput) ([]GitLabProject, error) {
	normalized, err := normalizeGitLabSearchInput(input)
	if err != nil {
		return nil, err
	}
	return s.client.WithAuth(normalized.APIBaseURL, normalized.AccessToken).SearchProjects(ctx, GitLabSearchOptions{
		Keyword: normalized.Keyword,
		Page:    normalized.Page,
		PerPage: normalized.Size,
	})
}

func (s *GitLabService) SearchGroups(ctx context.Context, input GitLabSearchInput) ([]GitLabGroup, error) {
	normalized, err := normalizeGitLabSearchInput(input)
	if err != nil {
		return nil, err
	}
	return s.client.WithAuth(normalized.APIBaseURL, normalized.AccessToken).SearchGroups(ctx, GitLabSearchOptions{
		Keyword: normalized.Keyword,
		Page:    normalized.Page,
		PerPage: normalized.Size,
	})
}

func normalizeGitLabSearchInput(input GitLabSearchInput) (GitLabSearchInput, error) {
	input.APIBaseURL = strings.TrimRight(strings.TrimSpace(input.APIBaseURL), "/")
	input.AccessToken = strings.TrimSpace(input.AccessToken)
	input.Keyword = strings.TrimSpace(input.Keyword)
	if input.APIBaseURL == "" || input.AccessToken == "" {
		return GitLabSearchInput{}, ErrInvalidGitLabInput
	}
	if _, err := url.ParseRequestURI(input.APIBaseURL); err != nil {
		return GitLabSearchInput{}, ErrInvalidGitLabInput
	}
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.Size <= 0 {
		input.Size = 20
	}
	if input.Size > 100 {
		input.Size = 100
	}
	return input, nil
}
