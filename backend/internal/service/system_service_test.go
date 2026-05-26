package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemServiceReturnsDefaultConfigWhenSettingMissing(t *testing.T) {
	svc := NewSystemService(&fakeSettingRepository{})

	config, err := svc.GetConfig(context.Background())

	require.NoError(t, err)
	require.Equal(t, "1.0.0", config.Version)
	require.Equal(t, "AI Code Review", config.SiteName)
	require.Equal(t, "", config.SiteNotice)
	require.Equal(t, "", config.BaseURL)
}

func TestSystemServiceParsesPersistedConfig(t *testing.T) {
	repo := &fakeSettingRepository{
		value: `{"version":"2.0.0","siteName":"Review Hub","siteNotice":"maintenance","baseUrl":"https://cr.example.com"}`,
		found: true,
	}
	svc := NewSystemService(repo)

	config, err := svc.GetConfig(context.Background())

	require.NoError(t, err)
	require.Equal(t, "2.0.0", config.Version)
	require.Equal(t, "Review Hub", config.SiteName)
	require.Equal(t, "maintenance", config.SiteNotice)
	require.Equal(t, "https://cr.example.com", config.BaseURL)
}

func TestSystemServiceReturnsErrorForInvalidPersistedJSON(t *testing.T) {
	repo := &fakeSettingRepository{value: `{invalid`, found: true}
	svc := NewSystemService(repo)

	_, err := svc.GetConfig(context.Background())

	require.Error(t, err)
}

func TestSystemServiceUpdateBaseURLValidatesAndPersists(t *testing.T) {
	repo := &fakeSettingRepository{
		value: `{"version":"2.0.0","siteName":"Review Hub","siteNotice":"notice"}`,
		found: true,
	}
	svc := NewSystemService(repo)

	config, err := svc.UpdateBaseURL(context.Background(), " https://cr.example.com/ ")

	require.NoError(t, err)
	require.Equal(t, "https://cr.example.com", config.BaseURL)
	require.JSONEq(t, `{"version":"2.0.0","siteName":"Review Hub","siteNotice":"notice","baseUrl":"https://cr.example.com"}`, repo.savedValue)
}

func TestSystemServiceUpdateBaseURLRejectsInvalidURL(t *testing.T) {
	svc := NewSystemService(&fakeSettingRepository{})

	tests := []string{"", "   ", "/relative", "ftp://example.com", "https:///missing-host"}
	for _, baseURL := range tests {
		t.Run(baseURL, func(t *testing.T) {
			_, err := svc.UpdateBaseURL(context.Background(), baseURL)
			require.ErrorIs(t, err, ErrInvalidSystemConfigInput)
		})
	}
}

type fakeSettingRepository struct {
	value      string
	found      bool
	savedKey   string
	savedValue string
	err        error
}

func (r *fakeSettingRepository) GetSettingValue(ctx context.Context, key string) (string, bool, error) {
	if r.err != nil {
		return "", false, r.err
	}
	if key != systemConfigSettingKey {
		return "", false, errors.New("unexpected key")
	}
	return r.value, r.found, nil
}

func (r *fakeSettingRepository) SetSettingValue(ctx context.Context, key string, value string) error {
	if r.err != nil {
		return r.err
	}
	r.savedKey = key
	r.savedValue = value
	return nil
}
