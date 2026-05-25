package database

import (
	"testing"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/config"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestMySQLDSNSupportsSpecialCharactersInPassword(t *testing.T) {
	dsn := MySQLDSN(config.DatabaseConfig{
		Host:      "127.0.0.1",
		Port:      3306,
		Username:  "review",
		Password:  "pa:ss@word",
		DBName:    "ai_review",
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	})

	parsed, err := mysql.ParseDSN(dsn)

	require.NoError(t, err)
	require.Equal(t, "review", parsed.User)
	require.Equal(t, "pa:ss@word", parsed.Passwd)
	require.Equal(t, "127.0.0.1:3306", parsed.Addr)
	require.Equal(t, "ai_review", parsed.DBName)
}

