package gitlab

import (
	"context"
	"net/http"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
)

type ServiceAdapter struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewServiceAdapter(client *http.Client) *ServiceAdapter {
	return &ServiceAdapter{client: client}
}

func (a *ServiceAdapter) WithAuth(baseURL string, token string) service.GitLabClient {
	return &ServiceAdapter{
		baseURL: baseURL,
		token:   token,
		client:  a.client,
	}
}

func (a *ServiceAdapter) SearchProjects(ctx context.Context, opts service.GitLabSearchOptions) ([]service.GitLabProject, error) {
	projects, err := NewClient(a.baseURL, a.token, a.client).SearchProjects(ctx, SearchOptions{
		Keyword: opts.Keyword,
		Page:    opts.Page,
		PerPage: opts.PerPage,
	})
	if err != nil {
		return nil, err
	}
	out := make([]service.GitLabProject, 0, len(projects))
	for _, project := range projects {
		out = append(out, service.GitLabProject{
			ID:                project.ID,
			Name:              project.Name,
			PathWithNamespace: project.PathWithNamespace,
			WebURL:            project.WebURL,
			DefaultBranch:     project.DefaultBranch,
		})
	}
	return out, nil
}

func (a *ServiceAdapter) SearchGroups(ctx context.Context, opts service.GitLabSearchOptions) ([]service.GitLabGroup, error) {
	groups, err := NewClient(a.baseURL, a.token, a.client).SearchGroups(ctx, SearchOptions{
		Keyword: opts.Keyword,
		Page:    opts.Page,
		PerPage: opts.PerPage,
	})
	if err != nil {
		return nil, err
	}
	out := make([]service.GitLabGroup, 0, len(groups))
	for _, group := range groups {
		out = append(out, service.GitLabGroup{
			ID:       group.ID,
			Name:     group.Name,
			FullPath: group.FullPath,
			WebURL:   group.WebURL,
		})
	}
	return out, nil
}

func (a *ServiceAdapter) GetMergeRequestChanges(ctx context.Context, projectID int, mergeRequestIID int) ([]service.GitLabDiff, error) {
	diff, err := NewClient(a.baseURL, a.token, a.client).GetMergeRequestChanges(ctx, projectID, mergeRequestIID)
	if err != nil {
		return nil, err
	}
	return mapDiff(diff), nil
}

func (a *ServiceAdapter) GetCommitDiff(ctx context.Context, projectID int, sha string) ([]service.GitLabDiff, error) {
	diff, err := NewClient(a.baseURL, a.token, a.client).GetCommitDiff(ctx, projectID, sha)
	if err != nil {
		return nil, err
	}
	return mapDiff(diff), nil
}

func mapDiff(items []Diff) []service.GitLabDiff {
	out := make([]service.GitLabDiff, 0, len(items))
	for _, item := range items {
		out = append(out, service.GitLabDiff{
			OldPath: item.OldPath,
			NewPath: item.NewPath,
			Diff:    item.Diff,
		})
	}
	return out
}
