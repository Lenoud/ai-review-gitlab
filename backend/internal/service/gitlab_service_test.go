package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitLabServiceRemoteSearchDelegatesToClient(t *testing.T) {
	client := &fakeGitLabClient{
		projects: []GitLabProject{{ID: 1, Name: "ai-review", WebURL: "https://gitlab.example.com/group/ai-review"}},
	}
	svc := NewGitLabService(client)

	projects, err := svc.SearchProjects(context.Background(), GitLabSearchInput{
		APIBaseURL:  "https://gitlab.example.com",
		AccessToken: "secret",
		Keyword:     "ai",
		Page:        1,
		Size:        20,
	})

	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, "https://gitlab.example.com", client.lastBaseURL)
	require.Equal(t, "secret", client.lastToken)
	require.Equal(t, "ai", client.lastKeyword)
}

func TestGitLabServiceGroupSearchValidatesInput(t *testing.T) {
	svc := NewGitLabService(&fakeGitLabClient{})

	_, err := svc.SearchGroups(context.Background(), GitLabSearchInput{})

	require.ErrorIs(t, err, ErrInvalidGitLabInput)
}

type fakeGitLabClient struct {
	lastBaseURL string
	lastToken   string
	lastKeyword string
	projects    []GitLabProject
	groups      []GitLabGroup
}

func (c *fakeGitLabClient) WithAuth(baseURL string, token string) GitLabClient {
	c.lastBaseURL = baseURL
	c.lastToken = token
	return c
}

func (c *fakeGitLabClient) SearchProjects(ctx context.Context, opts GitLabSearchOptions) ([]GitLabProject, error) {
	c.lastKeyword = opts.Keyword
	return c.projects, nil
}

func (c *fakeGitLabClient) SearchGroups(ctx context.Context, opts GitLabSearchOptions) ([]GitLabGroup, error) {
	c.lastKeyword = opts.Keyword
	return c.groups, nil
}

func (c *fakeGitLabClient) GetMergeRequestChanges(ctx context.Context, projectID int, mergeRequestIID int) ([]GitLabDiff, error) {
	return nil, nil
}

func (c *fakeGitLabClient) GetCommitDiff(ctx context.Context, projectID int, sha string) ([]GitLabDiff, error) {
	return nil, nil
}
