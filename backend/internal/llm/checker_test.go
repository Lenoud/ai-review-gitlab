package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/stretchr/testify/require"
)

func TestOpenAICompatibleCheckerSucceeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer secret", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer server.Close()

	checker := NewOpenAICompatibleChecker(server.Client())
	err := checker.Check(context.Background(), service.LLMConnectionInput{
		Provider:   "openai",
		ModelCode:  "gpt-4o-mini",
		APIBaseURL: server.URL + "/v1",
		APIKey:     "secret",
	})

	require.NoError(t, err)
}

func TestOpenAICompatibleCheckerReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad key", http.StatusUnauthorized)
	}))
	defer server.Close()

	checker := NewOpenAICompatibleChecker(server.Client())
	err := checker.Check(context.Background(), service.LLMConnectionInput{
		Provider:   "openai",
		ModelCode:  "gpt-4o-mini",
		APIBaseURL: server.URL + "/v1",
		APIKey:     "bad",
	})

	require.Error(t, err)
}
