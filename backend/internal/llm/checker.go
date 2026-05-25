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

type OpenAICompatibleChecker struct {
	client *http.Client
}

func NewOpenAICompatibleChecker(client *http.Client) *OpenAICompatibleChecker {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &OpenAICompatibleChecker{client: client}
}

func (c *OpenAICompatibleChecker) Check(ctx context.Context, input service.LLMConnectionInput) error {
	body, err := json.Marshal(map[string]any{
		"model": input.ModelCode,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
		"max_tokens": 1,
	})
	if err != nil {
		return err
	}

	endpoint := strings.TrimRight(input.APIBaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+input.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("llm connection failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	return nil
}
