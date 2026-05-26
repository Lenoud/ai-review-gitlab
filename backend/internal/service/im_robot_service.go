package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	IMRobotPlatformDingTalk = "dingtalk"
	IMRobotPlatformFeishu   = "feishu"
	IMRobotPlatformWeCom    = "wecom"

	imRobotPlatformMaxLength   = 32
	imRobotNameMaxLength       = 128
	imRobotWebhookURLMaxLength = 1024
	imRobotSecretMaxLength     = 512
)

var (
	ErrIMRobotNotFound     = errors.New("im robot not found")
	ErrInvalidIMRobotInput = errors.New("invalid im robot input")
	ErrIMRobotInUse        = errors.New("im robot in use")
)

type IMRobot struct {
	ID         uint   `json:"id"`
	Platform   string `json:"platform"`
	Name       string `json:"name"`
	WebhookURL string `json:"webhookUrl"`
	Secret     string `json:"secret,omitempty"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

type IMRobotInput struct {
	Platform   string
	Name       string
	WebhookURL string
	Secret     string
	Enabled    bool
	EnabledSet bool
}

type IMRobotSearchQuery struct {
	Keyword  string
	Platform string
	Enabled  *bool
	Page     int
	Size     int
}

type IMRobotPage struct {
	Items []IMRobot `json:"items"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}

type IMRobotTestWebhookInput struct {
	Platform   string
	WebhookURL string
}

type IMRobotTestWebhookResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type IMRobotRepository interface {
	CreateIMRobot(ctx context.Context, input IMRobotInput) (*IMRobot, error)
	UpdateIMRobot(ctx context.Context, id uint, input IMRobotInput) (*IMRobot, error)
	FindIMRobotByID(ctx context.Context, id uint) (*IMRobot, error)
	DeleteIMRobots(ctx context.Context, ids []uint) error
	SearchIMRobots(ctx context.Context, query IMRobotSearchQuery) (*IMRobotPage, error)
	ListEnabledIMRobots(ctx context.Context) ([]IMRobot, error)
	CountIMRobotReferences(ctx context.Context, ids []uint) (int64, error)
}

type IMRobotWebhookSender interface {
	SendIMRobotWebhook(ctx context.Context, webhookURL string, payload []byte) ([]byte, error)
}

type IMRobotService struct {
	robots        IMRobotRepository
	webhookSender IMRobotWebhookSender
}

func NewIMRobotService(robots IMRobotRepository) *IMRobotService {
	return NewIMRobotServiceWithSender(robots, newHTTPIMRobotWebhookSender(nil))
}

func NewIMRobotServiceWithSender(robots IMRobotRepository, webhookSender IMRobotWebhookSender) *IMRobotService {
	if webhookSender == nil {
		webhookSender = newHTTPIMRobotWebhookSender(nil)
	}
	return &IMRobotService{robots: robots, webhookSender: webhookSender}
}

func (s *IMRobotService) Create(ctx context.Context, input IMRobotInput) (*IMRobot, error) {
	normalized, err := normalizeIMRobotInput(input)
	if err != nil {
		return nil, err
	}
	return s.robots.CreateIMRobot(ctx, normalized)
}

func (s *IMRobotService) Update(ctx context.Context, id uint, input IMRobotInput) (*IMRobot, error) {
	if id == 0 {
		return nil, ErrInvalidIMRobotInput
	}
	normalized, err := normalizeIMRobotInput(input)
	if err != nil {
		return nil, err
	}
	return s.robots.UpdateIMRobot(ctx, id, normalized)
}

func (s *IMRobotService) Get(ctx context.Context, id uint) (*IMRobot, error) {
	if id == 0 {
		return nil, ErrInvalidIMRobotInput
	}
	return s.robots.FindIMRobotByID(ctx, id)
}

func (s *IMRobotService) Delete(ctx context.Context, ids []uint) error {
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
		return ErrInvalidIMRobotInput
	}
	count, err := s.robots.CountIMRobotReferences(ctx, cleanIDs)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrIMRobotInUse
	}
	return s.robots.DeleteIMRobots(ctx, cleanIDs)
}

func (s *IMRobotService) Search(ctx context.Context, query IMRobotSearchQuery) (*IMRobotPage, error) {
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
	return s.robots.SearchIMRobots(ctx, query)
}

func (s *IMRobotService) ListEnabled(ctx context.Context) ([]IMRobot, error) {
	return s.robots.ListEnabledIMRobots(ctx)
}

func (s *IMRobotService) TestWebhook(ctx context.Context, input IMRobotTestWebhookInput) (*IMRobotTestWebhookResult, error) {
	normalized, err := normalizeIMRobotTestWebhookInput(input)
	if err != nil {
		return nil, err
	}
	payload, err := buildIMRobotTestWebhookPayload(normalized.Platform)
	if err != nil {
		return nil, err
	}
	body, err := s.webhookSender.SendIMRobotWebhook(ctx, normalized.WebhookURL, payload)
	if err != nil {
		return &IMRobotTestWebhookResult{Success: false, Message: "请求异常: " + err.Error()}, nil
	}
	return parseIMRobotWebhookResponse(body), nil
}

func normalizeIMRobotInput(input IMRobotInput) (IMRobotInput, error) {
	input.Platform = strings.TrimSpace(input.Platform)
	input.Name = strings.TrimSpace(input.Name)
	input.WebhookURL = strings.TrimSpace(input.WebhookURL)
	input.Secret = strings.TrimSpace(input.Secret)
	if input.Platform == "" || input.Name == "" || input.WebhookURL == "" {
		return IMRobotInput{}, ErrInvalidIMRobotInput
	}
	if len(input.Platform) > imRobotPlatformMaxLength ||
		len(input.Name) > imRobotNameMaxLength ||
		len(input.WebhookURL) > imRobotWebhookURLMaxLength ||
		len(input.Secret) > imRobotSecretMaxLength {
		return IMRobotInput{}, ErrInvalidIMRobotInput
	}
	if !isSupportedIMRobotPlatform(input.Platform) {
		return IMRobotInput{}, ErrInvalidIMRobotInput
	}
	parsed, err := url.ParseRequestURI(input.WebhookURL)
	if err != nil || parsed.Host == "" || !isHTTPWebhookScheme(parsed.Scheme) {
		return IMRobotInput{}, ErrInvalidIMRobotInput
	}
	if !input.EnabledSet {
		input.Enabled = true
		input.EnabledSet = true
	}
	return input, nil
}

func normalizeIMRobotTestWebhookInput(input IMRobotTestWebhookInput) (IMRobotTestWebhookInput, error) {
	input.Platform = strings.TrimSpace(input.Platform)
	input.WebhookURL = strings.TrimSpace(input.WebhookURL)
	if input.Platform == "" || input.WebhookURL == "" {
		return IMRobotTestWebhookInput{}, ErrInvalidIMRobotInput
	}
	if len(input.Platform) > imRobotPlatformMaxLength || len(input.WebhookURL) > imRobotWebhookURLMaxLength {
		return IMRobotTestWebhookInput{}, ErrInvalidIMRobotInput
	}
	if !isSupportedIMRobotPlatform(input.Platform) {
		return IMRobotTestWebhookInput{}, ErrInvalidIMRobotInput
	}
	parsed, err := url.ParseRequestURI(input.WebhookURL)
	if err != nil || parsed.Host == "" || !isHTTPWebhookScheme(parsed.Scheme) {
		return IMRobotTestWebhookInput{}, ErrInvalidIMRobotInput
	}
	return input, nil
}

func buildIMRobotTestWebhookPayload(platform string) ([]byte, error) {
	switch platform {
	case IMRobotPlatformDingTalk, IMRobotPlatformWeCom:
		return []byte(`{"msgtype":"text","text":{"content":"Webhook 连接测试成功"}}`), nil
	case IMRobotPlatformFeishu:
		return []byte(`{"msg_type":"text","content":{"text":"Webhook 连接测试成功"}}`), nil
	default:
		return nil, ErrInvalidIMRobotInput
	}
}

func parseIMRobotWebhookResponse(body []byte) *IMRobotTestWebhookResult {
	if strings.TrimSpace(string(body)) == "" {
		return &IMRobotTestWebhookResult{Success: false, Message: "响应为空"}
	}
	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return &IMRobotTestWebhookResult{Success: false, Message: "响应格式异常，无法解析"}
	}
	if code, ok := optionalWebhookInt(root, "errcode"); ok {
		message := optionalWebhookText(root, "errmsg")
		if code != 0 {
			if message == "" {
				message = "errcode=" + strconv.Itoa(code)
			}
			return &IMRobotTestWebhookResult{Success: false, Message: message}
		}
		if message == "" {
			message = "ok"
		}
		return &IMRobotTestWebhookResult{Success: true, Message: message}
	}
	if code, ok := optionalWebhookInt(root, "code"); ok {
		message := optionalWebhookText(root, "msg")
		if code != 0 {
			if message == "" {
				message = "code=" + strconv.Itoa(code)
			}
			return &IMRobotTestWebhookResult{Success: false, Message: message}
		}
		if message == "" {
			message = "ok"
		}
		return &IMRobotTestWebhookResult{Success: true, Message: message}
	}
	return &IMRobotTestWebhookResult{Success: false, Message: "无法识别的响应格式"}
}

func optionalWebhookInt(root map[string]any, field string) (int, bool) {
	switch value := root[field].(type) {
	case float64:
		return int(value), true
	case string:
		parsed := 0
		for _, ch := range value {
			if ch < '0' || ch > '9' {
				return 0, false
			}
			parsed = parsed*10 + int(ch-'0')
		}
		return parsed, value != ""
	default:
		return 0, false
	}
}

func optionalWebhookText(root map[string]any, field string) string {
	value, ok := root[field].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func isHTTPWebhookScheme(scheme string) bool {
	return scheme == "http" || scheme == "https"
}

func isSupportedIMRobotPlatform(platform string) bool {
	switch platform {
	case IMRobotPlatformDingTalk, IMRobotPlatformFeishu, IMRobotPlatformWeCom:
		return true
	default:
		return false
	}
}

type httpIMRobotWebhookSender struct {
	client *http.Client
}

func newHTTPIMRobotWebhookSender(client *http.Client) *httpIMRobotWebhookSender {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &httpIMRobotWebhookSender{client: client}
}

func (s *httpIMRobotWebhookSender) SendIMRobotWebhook(ctx context.Context, webhookURL string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
