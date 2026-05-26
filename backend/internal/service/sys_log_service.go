package service

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrSysLogNotFound     = errors.New("sys log not found")
	ErrInvalidSysLogInput = errors.New("invalid sys log input")
)

type SysLog struct {
	ID         uint   `json:"id"`
	Level      string `json:"level"`
	Module     string `json:"module"`
	Action     string `json:"action"`
	Message    string `json:"message"`
	Detail     string `json:"detail"`
	ErrorStack string `json:"errorStack"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

type SysLogSearchQuery struct {
	Level     string
	Module    string
	Action    string
	Message   string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	Size      int
}

type SysLogPage struct {
	Items []SysLog `json:"items"`
	Total int64    `json:"total"`
	Page  int      `json:"page"`
	Size  int      `json:"size"`
}

type SysLogRepository interface {
	FindSysLogByID(ctx context.Context, id uint) (*SysLog, error)
	SearchSysLogs(ctx context.Context, query SysLogSearchQuery) (*SysLogPage, error)
}

type SysLogService struct {
	logs SysLogRepository
}

func NewSysLogService(logs SysLogRepository) *SysLogService {
	return &SysLogService{logs: logs}
}

func (s *SysLogService) Get(ctx context.Context, id uint) (*SysLog, error) {
	if id == 0 {
		return nil, ErrInvalidSysLogInput
	}
	return s.logs.FindSysLogByID(ctx, id)
}

func (s *SysLogService) Search(ctx context.Context, query SysLogSearchQuery) (*SysLogPage, error) {
	query.Level = strings.ToUpper(strings.TrimSpace(query.Level))
	query.Module = strings.ToUpper(strings.TrimSpace(query.Module))
	query.Action = strings.TrimSpace(query.Action)
	query.Message = strings.TrimSpace(query.Message)
	query.Page, query.Size = normalizeStatsPage(query.Page, query.Size)
	return s.logs.SearchSysLogs(ctx, query)
}
