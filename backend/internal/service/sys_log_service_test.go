package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSysLogServiceGetValidatesIDAndDelegates(t *testing.T) {
	repo := &fakeSysLogRepository{log: SysLog{ID: 7, Level: "INFO", Module: "REVIEW"}}
	got, err := NewSysLogService(repo).Get(context.Background(), 7)

	require.NoError(t, err)
	require.Equal(t, uint(7), got.ID)
	require.Equal(t, "INFO", got.Level)

	_, err = NewSysLogService(repo).Get(context.Background(), 0)
	require.ErrorIs(t, err, ErrInvalidSysLogInput)
}

func TestSysLogServiceSearchNormalizesQuery(t *testing.T) {
	start := time.UnixMilli(1000)
	end := time.UnixMilli(2000)
	repo := &fakeSysLogRepository{
		page: SysLogPage{Items: []SysLog{{ID: 1, Level: "ERROR"}}, Total: 1, Page: 1, Size: 20},
	}
	page, err := NewSysLogService(repo).Search(context.Background(), SysLogSearchQuery{
		Level:     " info ",
		Module:    " review ",
		Action:    " send ",
		Message:   " failed ",
		StartTime: &start,
		EndTime:   &end,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "INFO", repo.query.Level)
	require.Equal(t, "REVIEW", repo.query.Module)
	require.Equal(t, "send", repo.query.Action)
	require.Equal(t, 1, repo.query.Page)
	require.Equal(t, 20, repo.query.Size)
}

type fakeSysLogRepository struct {
	log   SysLog
	page  SysLogPage
	query SysLogSearchQuery
	err   error
}

func (r *fakeSysLogRepository) FindSysLogByID(ctx context.Context, id uint) (*SysLog, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &r.log, nil
}

func (r *fakeSysLogRepository) SearchSysLogs(ctx context.Context, query SysLogSearchQuery) (*SysLogPage, error) {
	if r.err != nil {
		return nil, r.err
	}
	r.query = query
	return &r.page, nil
}
