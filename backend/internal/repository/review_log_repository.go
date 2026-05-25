package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ReviewLogRepository struct {
	db *gorm.DB
}

func NewReviewLogRepository(db *gorm.DB) *ReviewLogRepository {
	return &ReviewLogRepository{db: db}
}

func (r *ReviewLogRepository) CreatePush(ctx context.Context, input service.PushReviewLogInput) (*service.PushReviewLog, error) {
	record, err := pushReviewLogModelFromInput(input)
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return pushReviewLogModelToService(record)
}

func (r *ReviewLogRepository) FindPushByID(ctx context.Context, id uint) (*service.PushReviewLog, error) {
	var record model.PushReviewLog
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrReviewLogNotFound
		}
		return nil, err
	}
	return pushReviewLogModelToService(&record)
}

func (r *ReviewLogRepository) SearchPush(ctx context.Context, query service.ReviewLogSearchQuery) (*service.PushReviewLogPage, error) {
	db := applyCommonReviewLogFilters(r.db.WithContext(ctx).Model(&model.PushReviewLog{}), query)
	if query.Branch != "" {
		db = db.Where("branch LIKE ?", "%"+query.Branch+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.PushReviewLog
	offset := (query.Page - 1) * query.Size
	if err := db.Order("id DESC").Offset(offset).Limit(query.Size).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.PushReviewLog, 0, len(records))
	for i := range records {
		item, err := pushReviewLogModelToService(&records[i])
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return &service.PushReviewLogPage{Items: items, Total: total, Page: query.Page, Size: query.Size}, nil
}

func (r *ReviewLogRepository) CreateMergeRequest(ctx context.Context, input service.MergeRequestReviewLogInput) (*service.MergeRequestReviewLog, error) {
	record := mergeRequestReviewLogModelFromInput(input)
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return mergeRequestReviewLogModelToService(record), nil
}

func (r *ReviewLogRepository) FindMergeRequestByID(ctx context.Context, id uint) (*service.MergeRequestReviewLog, error) {
	var record model.MergeRequestReviewLog
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrReviewLogNotFound
		}
		return nil, err
	}
	return mergeRequestReviewLogModelToService(&record), nil
}

func (r *ReviewLogRepository) SearchMergeRequest(ctx context.Context, query service.ReviewLogSearchQuery) (*service.MergeRequestReviewLogPage, error) {
	db := applyCommonReviewLogFilters(r.db.WithContext(ctx).Model(&model.MergeRequestReviewLog{}), query)
	if query.Branch != "" {
		like := "%" + query.Branch + "%"
		db = db.Where("source_branch LIKE ? OR target_branch LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.MergeRequestReviewLog
	offset := (query.Page - 1) * query.Size
	if err := db.Order("id DESC").Offset(offset).Limit(query.Size).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.MergeRequestReviewLog, 0, len(records))
	for i := range records {
		items = append(items, *mergeRequestReviewLogModelToService(&records[i]))
	}
	return &service.MergeRequestReviewLogPage{Items: items, Total: total, Page: query.Page, Size: query.Size}, nil
}

func applyCommonReviewLogFilters(db *gorm.DB, query service.ReviewLogSearchQuery) *gorm.DB {
	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
	if len(query.Authors) > 0 {
		db = db.Where("author_identity IN ?", query.Authors)
	}
	if len(query.ProjectNames) > 0 {
		db = db.Where("project_name IN ?", query.ProjectNames)
	}
	if query.MinScore != nil {
		db = db.Where("score >= ?", *query.MinScore)
	}
	if query.MaxScore != nil {
		db = db.Where("score <= ?", *query.MaxScore)
	}
	if query.CommitMessages != "" {
		db = db.Where("commit_messages LIKE ?", "%"+query.CommitMessages+"%")
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}
	return db
}

func pushReviewLogModelFromInput(input service.PushReviewLogInput) (*model.PushReviewLog, error) {
	commits, err := json.Marshal(input.Commits)
	if err != nil {
		return nil, err
	}
	return &model.PushReviewLog{
		ProjectID:         input.ProjectID,
		ProjectName:       input.ProjectName,
		Author:            input.Author,
		AuthorIdentity:    input.AuthorIdentity,
		AuthorDisplayName: input.AuthorDisplayName,
		Branch:            input.Branch,
		CommitMessages:    input.CommitMessages,
		Commits:           datatypes.JSON(commits),
		Score:             input.Score,
		Additions:         input.Additions,
		Deletions:         input.Deletions,
		LastCommitURL:     input.LastCommitURL,
		ReviewResult:      input.ReviewResult,
	}, nil
}

func mergeRequestReviewLogModelFromInput(input service.MergeRequestReviewLogInput) *model.MergeRequestReviewLog {
	return &model.MergeRequestReviewLog{
		ProjectID:         input.ProjectID,
		ProjectName:       input.ProjectName,
		Author:            input.Author,
		AuthorIdentity:    input.AuthorIdentity,
		AuthorDisplayName: input.AuthorDisplayName,
		SourceBranch:      input.SourceBranch,
		TargetBranch:      input.TargetBranch,
		CommitMessages:    input.CommitMessages,
		Score:             input.Score,
		Additions:         input.Additions,
		Deletions:         input.Deletions,
		LastCommitID:      input.LastCommitID,
		URL:               input.URL,
		ReviewResult:      input.ReviewResult,
	}
}

func pushReviewLogModelToService(record *model.PushReviewLog) (*service.PushReviewLog, error) {
	var commits []service.ReviewCommit
	if len(record.Commits) > 0 {
		if err := json.Unmarshal(record.Commits, &commits); err != nil {
			return nil, err
		}
	}
	return &service.PushReviewLog{
		ID:                  record.ID,
		ProjectID:           record.ProjectID,
		ProjectName:         record.ProjectName,
		Author:              record.Author,
		AuthorIdentity:      record.AuthorIdentity,
		AuthorDisplayName:   record.AuthorDisplayName,
		AuthorDisplayText:   authorDisplayText(record.AuthorIdentity, record.AuthorDisplayName),
		Branch:              record.Branch,
		CommitMessages:      record.CommitMessages,
		Commits:             commits,
		Score:               record.Score,
		Additions:           record.Additions,
		Deletions:           record.Deletions,
		LastCommitURL:       record.LastCommitURL,
		ReviewResult:        record.ReviewResult,
		ShareToken:          record.ShareToken,
		ShareTokenExpiresAt: record.ShareTokenExpiresAt,
		CreatedAt:           record.CreatedAt.UnixMilli(),
		UpdatedAt:           record.UpdatedAt.UnixMilli(),
	}, nil
}

func mergeRequestReviewLogModelToService(record *model.MergeRequestReviewLog) *service.MergeRequestReviewLog {
	return &service.MergeRequestReviewLog{
		ID:                  record.ID,
		ProjectID:           record.ProjectID,
		ProjectName:         record.ProjectName,
		Author:              record.Author,
		AuthorIdentity:      record.AuthorIdentity,
		AuthorDisplayName:   record.AuthorDisplayName,
		AuthorDisplayText:   authorDisplayText(record.AuthorIdentity, record.AuthorDisplayName),
		CommitMessages:      record.CommitMessages,
		Score:               record.Score,
		SourceBranch:        record.SourceBranch,
		TargetBranch:        record.TargetBranch,
		Additions:           record.Additions,
		Deletions:           record.Deletions,
		LastCommitID:        record.LastCommitID,
		URL:                 record.URL,
		ReviewResult:        record.ReviewResult,
		ShareToken:          record.ShareToken,
		ShareTokenExpiresAt: record.ShareTokenExpiresAt,
		CreatedAt:           record.CreatedAt.UnixMilli(),
		UpdatedAt:           record.UpdatedAt.UnixMilli(),
	}
}

func authorDisplayText(identity string, displayName string) string {
	identity = strings.TrimSpace(identity)
	displayName = strings.TrimSpace(displayName)
	switch {
	case identity == "":
		return displayName
	case displayName == "":
		return identity
	default:
		return identity + "（" + displayName + "）"
	}
}
