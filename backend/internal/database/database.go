package database

import (
	"fmt"
	"net/url"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/config"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := MySQLDSN(cfg)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func MySQLDSN(cfg config.DatabaseConfig) string {
	charset := cfg.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	loc := cfg.Loc
	if loc == "" {
		loc = "Local"
	}
	query := url.Values{}
	query.Set("charset", charset)
	query.Set("parseTime", fmt.Sprintf("%t", cfg.ParseTime))
	query.Set("loc", loc)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		query.Encode(),
	)
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.SysUser{},
		&model.SysRole{},
		&model.SysPermission{},
		&model.SysUserRole{},
		&model.SysRolePermission{},
		&model.Project{},
		&model.LLMModel{},
		&model.ReviewTask{},
		&model.ReviewTaskAttempt{},
		&model.PushReviewLog{},
		&model.MergeRequestReviewLog{},
		&model.ProjectAnalysisPlan{},
		&model.ProjectAnalysisPlanExecutionLog{},
		&model.AIReviewTrace{},
		&model.IMRobot{},
		&model.ProjectTemplate{},
		&model.ProjectTemplateReviewRule{},
	)
}
