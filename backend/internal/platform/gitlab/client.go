package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type SearchOptions struct {
	Keyword string
	Page    int
	PerPage int
}

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	DefaultBranch     string `json:"default_branch"`
}

type Group struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullPath string `json:"full_path"`
	WebURL   string `json:"web_url"`
}

type Diff struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
	Diff    string `json:"diff"`
}

type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewClient(baseURL string, token string, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:   strings.TrimSpace(token),
		client:  client,
	}
}

func (c *Client) SearchProjects(ctx context.Context, opts SearchOptions) ([]Project, error) {
	var projects []Project
	query := searchQuery(opts)
	query.Set("simple", "true")
	if err := c.get(ctx, "/api/v4/projects", query, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *Client) SearchGroups(ctx context.Context, opts SearchOptions) ([]Group, error) {
	var groups []Group
	if err := c.get(ctx, "/api/v4/groups", searchQuery(opts), &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (c *Client) GetMergeRequestChanges(ctx context.Context, projectID int, mergeRequestIID int) ([]Diff, error) {
	var response struct {
		Changes []Diff `json:"changes"`
	}
	endpoint := fmt.Sprintf("/api/v4/projects/%d/merge_requests/%d/changes", projectID, mergeRequestIID)
	if err := c.get(ctx, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response.Changes, nil
}

func (c *Client) GetCommitDiff(ctx context.Context, projectID int, sha string) ([]Diff, error) {
	var diff []Diff
	endpoint := fmt.Sprintf("/api/v4/projects/%d/repository/commits/%s/diff", projectID, url.PathEscape(sha))
	if err := c.get(ctx, endpoint, nil, &diff); err != nil {
		return nil, err
	}
	return diff, nil
}

func (c *Client) get(ctx context.Context, endpoint string, query url.Values, out any) error {
	if c.baseURL == "" {
		return fmt.Errorf("gitlab base url is required")
	}
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, endpoint)
	if query != nil {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("gitlab request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func searchQuery(opts SearchOptions) url.Values {
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	perPage := opts.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	query := url.Values{}
	query.Set("search", strings.TrimSpace(opts.Keyword))
	query.Set("page", strconv.Itoa(page))
	query.Set("per_page", strconv.Itoa(perPage))
	return query
}
