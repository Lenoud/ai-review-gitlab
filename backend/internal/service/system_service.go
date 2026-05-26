package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

const systemConfigSettingKey = "SYSTEM"

var ErrInvalidSystemConfigInput = errors.New("invalid system config input")

type SystemConfig struct {
	Version    string `json:"version"`
	SiteName   string `json:"siteName"`
	SiteNotice string `json:"siteNotice"`
	BaseURL    string `json:"baseUrl"`
}

type SystemConfigRepository interface {
	GetSettingValue(ctx context.Context, key string) (string, bool, error)
	SetSettingValue(ctx context.Context, key string, value string) error
}

type SystemService struct {
	settings SystemConfigRepository
}

func NewSystemService(settings SystemConfigRepository) *SystemService {
	return &SystemService{settings: settings}
}

func (s *SystemService) GetConfig(ctx context.Context) (*SystemConfig, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *SystemService) UpdateBaseURL(ctx context.Context, baseURL string) (*SystemConfig, error) {
	normalized, err := normalizeSystemBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	config.BaseURL = normalized
	value, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	if err := s.settings.SetSettingValue(ctx, systemConfigSettingKey, string(value)); err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *SystemService) loadConfig(ctx context.Context) (SystemConfig, error) {
	config := defaultSystemConfig()
	value, found, err := s.settings.GetSettingValue(ctx, systemConfigSettingKey)
	if err != nil {
		return SystemConfig{}, err
	}
	if !found {
		return config, nil
	}
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return SystemConfig{}, err
	}
	applySystemConfigDefaults(&config)
	return config, nil
}

func defaultSystemConfig() SystemConfig {
	return SystemConfig{
		Version:    "1.0.0",
		SiteName:   "AI Code Review",
		SiteNotice: "",
		BaseURL:    "",
	}
}

func applySystemConfigDefaults(config *SystemConfig) {
	if strings.TrimSpace(config.Version) == "" {
		config.Version = "1.0.0"
	}
	if strings.TrimSpace(config.SiteName) == "" {
		config.SiteName = "AI Code Review"
	}
}

func normalizeSystemBaseURL(baseURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return "", ErrInvalidSystemConfigInput
	}
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", ErrInvalidSystemConfigInput
	}
	return baseURL, nil
}
