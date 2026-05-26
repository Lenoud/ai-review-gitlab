package repository

import (
	"context"
	"errors"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

type SysLogRepository struct {
	db *gorm.DB
}

func NewSysLogRepository(db *gorm.DB) *SysLogRepository {
	return &SysLogRepository{db: db}
}

func (r *SysLogRepository) FindSysLogByID(ctx context.Context, id uint) (*service.SysLog, error) {
	var record model.SysLog
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrSysLogNotFound
		}
		return nil, err
	}
	return sysLogModelToService(&record), nil
}

func (r *SysLogRepository) SearchSysLogs(ctx context.Context, query service.SysLogSearchQuery) (*service.SysLogPage, error) {
	db := r.db.WithContext(ctx).Model(&model.SysLog{})
	if query.Level != "" {
		db = db.Where("level = ?", query.Level)
	}
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}
	if query.Action != "" {
		db = db.Where("action LIKE ?", "%"+query.Action+"%")
	}
	if query.Message != "" {
		db = db.Where("message LIKE ?", "%"+query.Message+"%")
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.SysLog
	offset := (query.Page - 1) * query.Size
	if err := db.Order("created_at DESC, id DESC").Offset(offset).Limit(query.Size).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.SysLog, 0, len(records))
	for i := range records {
		items = append(items, *sysLogModelToService(&records[i]))
	}
	return &service.SysLogPage{Items: items, Total: total, Page: query.Page, Size: query.Size}, nil
}

func sysLogModelToService(record *model.SysLog) *service.SysLog {
	return &service.SysLog{
		ID:         record.ID,
		Level:      record.Level,
		Module:     record.Module,
		Action:     record.Action,
		Message:    record.Message,
		Detail:     record.Detail,
		ErrorStack: record.ErrorStack,
		CreatedAt:  record.CreatedAt.UnixMilli(),
		UpdatedAt:  record.UpdatedAt.UnixMilli(),
	}
}
