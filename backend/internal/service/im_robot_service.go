package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
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

type IMRobotRepository interface {
	CreateIMRobot(ctx context.Context, input IMRobotInput) (*IMRobot, error)
	UpdateIMRobot(ctx context.Context, id uint, input IMRobotInput) (*IMRobot, error)
	FindIMRobotByID(ctx context.Context, id uint) (*IMRobot, error)
	DeleteIMRobots(ctx context.Context, ids []uint) error
	SearchIMRobots(ctx context.Context, query IMRobotSearchQuery) (*IMRobotPage, error)
	ListEnabledIMRobots(ctx context.Context) ([]IMRobot, error)
	CountIMRobotReferences(ctx context.Context, ids []uint) (int64, error)
}

type IMRobotService struct {
	robots IMRobotRepository
}

func NewIMRobotService(robots IMRobotRepository) *IMRobotService {
	return &IMRobotService{robots: robots}
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
