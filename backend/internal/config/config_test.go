package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, 8080, cfg.Server.Port)
	require.Equal(t, "0.0.0.0", cfg.Server.Host)
	require.Equal(t, "127.0.0.1", cfg.Database.Host)
	require.Equal(t, 3306, cfg.Database.Port)
	require.Equal(t, "ai_review", cfg.Database.DBName)
	require.Equal(t, "ai-review", cfg.Auth.Issuer)
	require.Equal(t, "admin", cfg.Auth.AdminUsername)
	require.Equal(t, "admin123", cfg.Auth.AdminPassword)
	require.Equal(t, "30m", cfg.Auth.AccessTokenTTL)
	require.Equal(t, "720h", cfg.Auth.RefreshTokenTTL)
	require.False(t, cfg.Worker.Enabled)
	require.Equal(t, "review-worker-1", cfg.Worker.ID)
	require.Equal(t, "5s", cfg.Worker.IdleInterval)
	require.Equal(t, "30s", cfg.Worker.ErrorInterval)
	require.Equal(t, 10000, cfg.Worker.MaxInputTokens)
}

func TestLoadEnvOverridesForDatabaseAndAdmin(t *testing.T) {
	t.Setenv("DATABASE_HOST", "mysql.internal")
	t.Setenv("DATABASE_PORT", "3307")
	t.Setenv("DATABASE_USERNAME", "review_user")
	t.Setenv("DATABASE_PASSWORD", "secret")
	t.Setenv("DATABASE_DBNAME", "review_db")
	t.Setenv("AUTH_ADMIN_USERNAME", "root")
	t.Setenv("AUTH_ADMIN_PASSWORD", "root-password")

	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, "mysql.internal", cfg.Database.Host)
	require.Equal(t, 3307, cfg.Database.Port)
	require.Equal(t, "review_user", cfg.Database.Username)
	require.Equal(t, "secret", cfg.Database.Password)
	require.Equal(t, "review_db", cfg.Database.DBName)
	require.Equal(t, "root", cfg.Auth.AdminUsername)
	require.Equal(t, "root-password", cfg.Auth.AdminPassword)
}

func TestLoadEnvOverridesForWorker(t *testing.T) {
	t.Setenv("WORKER_ENABLED", "true")
	t.Setenv("WORKER_ID", "worker-test")
	t.Setenv("WORKER_IDLE_INTERVAL", "1s")
	t.Setenv("WORKER_ERROR_INTERVAL", "2s")
	t.Setenv("WORKER_MAX_INPUT_TOKENS", "1234")

	cfg, err := Load()

	require.NoError(t, err)
	require.True(t, cfg.Worker.Enabled)
	require.Equal(t, "worker-test", cfg.Worker.ID)
	require.Equal(t, "1s", cfg.Worker.IdleInterval)
	require.Equal(t, "2s", cfg.Worker.ErrorInterval)
	require.Equal(t, 1234, cfg.Worker.MaxInputTokens)
}
