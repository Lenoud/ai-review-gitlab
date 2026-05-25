package gitlab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientSearchProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v4/projects", r.URL.Path)
		require.Equal(t, "secret", r.Header.Get("PRIVATE-TOKEN"))
		require.Equal(t, "ai-review", r.URL.Query().Get("search"))
		require.Equal(t, "1", r.URL.Query().Get("page"))
		require.Equal(t, "20", r.URL.Query().Get("per_page"))
		_, _ = w.Write([]byte(`[{"id":1,"name":"ai-review","path_with_namespace":"group/ai-review","web_url":"https://gitlab.example.com/group/ai-review","default_branch":"main"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret", server.Client())
	projects, err := client.SearchProjects(context.Background(), SearchOptions{Keyword: "ai-review", Page: 1, PerPage: 20})

	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, 1, projects[0].ID)
	require.Equal(t, "group/ai-review", projects[0].PathWithNamespace)
}

func TestClientSearchGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v4/groups", r.URL.Path)
		require.Equal(t, "platform", r.URL.Query().Get("search"))
		_, _ = w.Write([]byte(`[{"id":2,"name":"platform","full_path":"org/platform","web_url":"https://gitlab.example.com/org/platform"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret", server.Client())
	groups, err := client.SearchGroups(context.Background(), SearchOptions{Keyword: "platform", Page: 1, PerPage: 20})

	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Equal(t, "org/platform", groups[0].FullPath)
}

func TestClientGetMergeRequestChangesAndCommitDiff(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/projects/123/merge_requests/7/changes":
			_, _ = w.Write([]byte(`{"changes":[{"old_path":"a.go","new_path":"a.go","diff":"@@ diff"}]}`))
		case "/api/v4/projects/123/repository/commits/abc/diff":
			_, _ = w.Write([]byte(`[{"old_path":"b.go","new_path":"b.go","diff":"@@ commit"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "secret", server.Client())
	changes, err := client.GetMergeRequestChanges(context.Background(), 123, 7)
	require.NoError(t, err)
	require.Equal(t, "@@ diff", changes[0].Diff)

	diff, err := client.GetCommitDiff(context.Background(), 123, "abc")
	require.NoError(t, err)
	require.Equal(t, "@@ commit", diff[0].Diff)
}

func TestClientReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad token", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad", server.Client())
	_, err := client.SearchProjects(context.Background(), SearchOptions{Keyword: "x"})

	require.Error(t, err)
}
