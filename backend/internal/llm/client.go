package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
)

type OpenAICompatibleClient struct {
	client *http.Client
}

func NewOpenAICompatibleClient(client *http.Client) *OpenAICompatibleClient {
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &OpenAICompatibleClient{client: client}
}

func (c *OpenAICompatibleClient) Chat(ctx context.Context, input service.LLMChatInput) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":      strings.TrimSpace(input.ModelCode),
		"messages":   input.Messages,
		"max_tokens": normalizeMaxTokens(input.MaxTokens),
	})
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(strings.TrimSpace(input.APIBaseURL), "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(input.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("llm chat failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("llm chat failed: empty choices")
	}
	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("llm chat failed: empty content")
	}
	return content, nil
}

func normalizeMaxTokens(maxTokens int) int {
	if maxTokens <= 0 {
		return 4096
	}
	return maxTokens
}
