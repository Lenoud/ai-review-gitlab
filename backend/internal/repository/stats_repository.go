package repository

import (
	"context"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) ListReviewStatsEntries(ctx context.Context, query service.StatsEntryQuery) ([]service.ReviewStatsLogEntry, error) {
	pushRows, err := r.listPushStatsEntries(ctx, query)
	if err != nil {
		return nil, err
	}
	mergeRows, err := r.listMergeRequestStatsEntries(ctx, query)
	if err != nil {
		return nil, err
	}
	rows := make([]service.ReviewStatsLogEntry, 0, len(pushRows)+len(mergeRows))
	rows = append(rows, pushRows...)
	rows = append(rows, mergeRows...)
	return rows, nil
}

func (r *StatsRepository) listPushStatsEntries(ctx context.Context, query service.StatsEntryQuery) ([]service.ReviewStatsLogEntry, error) {
	var records []model.PushReviewLog
	db := r.db.WithContext(ctx).
		Model(&model.PushReviewLog{}).
		Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	if query.Project != "" {
		db = db.Where("project_name = ?", query.Project)
	}
	if err := db.Order("created_at ASC, id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	rows := make([]service.ReviewStatsLogEntry, 0, len(records))
	for _, record := range records {
		rows = append(rows, service.ReviewStatsLogEntry{
			ProjectID:         record.ProjectID,
			ProjectName:       record.ProjectName,
			Author:            record.Author,
			AuthorIdentity:    record.AuthorIdentity,
			AuthorDisplayName: record.AuthorDisplayName,
			Score:             record.Score,
			Additions:         record.Additions,
			Deletions:         record.Deletions,
			CreatedAt:         record.CreatedAt.UnixMilli(),
		})
	}
	return rows, nil
}

func (r *StatsRepository) listMergeRequestStatsEntries(ctx context.Context, query service.StatsEntryQuery) ([]service.ReviewStatsLogEntry, error) {
	var records []model.MergeRequestReviewLog
	db := r.db.WithContext(ctx).
		Model(&model.MergeRequestReviewLog{}).
		Where("created_at >= ? AND created_at <= ?", query.StartTime, query.EndTime)
	if query.Project != "" {
		db = db.Where("project_name = ?", query.Project)
	}
	if err := db.Order("created_at ASC, id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	rows := make([]service.ReviewStatsLogEntry, 0, len(records))
	for _, record := range records {
		rows = append(rows, service.ReviewStatsLogEntry{
			ProjectID:         record.ProjectID,
			ProjectName:       record.ProjectName,
			Author:            record.Author,
			AuthorIdentity:    record.AuthorIdentity,
			AuthorDisplayName: record.AuthorDisplayName,
			Score:             record.Score,
			Additions:         record.Additions,
			Deletions:         record.Deletions,
			CreatedAt:         record.CreatedAt.UnixMilli(),
		})
	}
	return rows, nil
}
