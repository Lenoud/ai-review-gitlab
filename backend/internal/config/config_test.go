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
	require.Equal(t, "ai-review", cfg.Auth.Issuer)
}
