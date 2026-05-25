package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"github.com/stretchr/testify/require"
)

func TestOpenAICompatibleClientChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer llm-key", r.Header.Get("Authorization"))
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "gpt-test", body["model"])
		require.Equal(t, float64(512), body["max_tokens"])
		messages := body["messages"].([]any)
		require.Len(t, messages, 1)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"review ok"}}]}`))
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(server.Client())
	content, err := client.Chat(context.Background(), service.LLMChatInput{
		APIBaseURL: server.URL + "/v1",
		APIKey:     "llm-key",
		ModelCode:  "gpt-test",
		MaxTokens:  512,
		Messages: []service.LLMChatMessage{
			{Role: "user", Content: "review this"},
		},
	})

	require.NoError(t, err)
	require.Equal(t, "review ok", content)
}

func TestOpenAICompatibleClientChatReturnsErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewOpenAICompatibleClient(server.Client())
	_, err := client.Chat(context.Background(), service.LLMChatInput{
		APIBaseURL: server.URL,
		APIKey:     "llm-key",
		ModelCode:  "gpt-test",
		Messages:   []service.LLMChatMessage{{Role: "user", Content: "hello"}},
	})

	require.Error(t, err)
}
