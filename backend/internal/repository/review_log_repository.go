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

func (r *ReviewLogRepository) DeletePush(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.PushReviewLog{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrReviewLogNotFound
	}
	return nil
}

func (r *ReviewLogRepository) DistinctPushAuthors(ctx context.Context, query service.ReviewLogOptionQuery) ([]service.AuthorOption, error) {
	var rows []authorRow
	db := applyOptionFilters(r.db.WithContext(ctx).Model(&model.PushReviewLog{}), query)
	if err := db.Select("author_identity, author_display_name").Where("author_identity <> ''").Order("author_identity ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mergeAuthorOptions(rows), nil
}

func (r *ReviewLogRepository) DistinctPushProjectNames(ctx context.Context, query service.ReviewLogOptionQuery) ([]string, error) {
	var names []string
	db := applyOptionFilters(r.db.WithContext(ctx).Model(&model.PushReviewLog{}), query)
	if err := db.Distinct("project_name").Where("project_name <> ''").Order("project_name ASC").Pluck("project_name", &names).Error; err != nil {
		return nil, err
	}
	return names, nil
}

func (r *ReviewLogRepository) UpdatePushShareToken(ctx context.Context, id uint, token string, expiresAt int64) (*service.PushReviewLog, error) {
	result := r.db.WithContext(ctx).Model(&model.PushReviewLog{}).Where("id = ?", id).Updates(map[string]any{
		"share_token":            token,
		"share_token_expires_at": expiresAt,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, service.ErrReviewLogNotFound
	}
	return r.FindPushByID(ctx, id)
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

func (r *ReviewLogRepository) FindAnalysisExecutionByID(ctx context.Context, id uint) (*service.ProjectAnalysisPlanExecutionLog, error) {
	var record model.ProjectAnalysisPlanExecutionLog
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrReviewLogNotFound
		}
		return nil, err
	}
	return analysisExecutionLogModelToService(&record), nil
}

func (r *ReviewLogRepository) CreateAnalysisExecution(ctx context.Context, input service.AnalysisExecutionLogInput) (*service.ProjectAnalysisPlanExecutionLog, error) {
	record := model.ProjectAnalysisPlanExecutionLog{
		PlanID:        input.PlanID,
		ProjectID:     input.ProjectID,
		Status:        input.Status,
		StartedAt:     input.StartedAt,
		CompletedAt:   input.CompletedAt,
		DurationMs:    input.DurationMs,
		ResultContent: input.ResultContent,
		ResultActions: input.ResultActions,
		ErrorMessage:  input.ErrorMessage,
		ErrorStack:    input.ErrorStack,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, err
	}
	return analysisExecutionLogModelToService(&record), nil
}

func (r *ReviewLogRepository) SearchAnalysisExecution(ctx context.Context, query service.AnalysisExecutionLogSearchQuery) (*service.AnalysisExecutionLogPage, error) {
	db := r.db.WithContext(ctx).Model(&model.ProjectAnalysisPlanExecutionLog{})
	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
	if query.PlanID > 0 {
		db = db.Where("plan_id = ?", query.PlanID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
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

	var records []model.ProjectAnalysisPlanExecutionLog
	offset := (query.Page - 1) * query.Size
	if err := db.Order("id DESC").Offset(offset).Limit(query.Size).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.ProjectAnalysisPlanExecutionLog, 0, len(records))
	for i := range records {
		items = append(items, *analysisExecutionLogModelToService(&records[i]))
	}
	return &service.AnalysisExecutionLogPage{Items: items, Total: total, Page: query.Page, Size: query.Size}, nil
}

func (r *ReviewLogRepository) UpdateAnalysisExecutionShareToken(ctx context.Context, id uint, token string, expiresAt int64) (*service.ProjectAnalysisPlanExecutionLog, error) {
	result := r.db.WithContext(ctx).Model(&model.ProjectAnalysisPlanExecutionLog{}).Where("id = ?", id).Updates(map[string]any{
		"share_token":            token,
		"share_token_expires_at": expiresAt,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, service.ErrReviewLogNotFound
	}
	return r.FindAnalysisExecutionByID(ctx, id)
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

func (r *ReviewLogRepository) DeleteMergeRequest(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.MergeRequestReviewLog{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrReviewLogNotFound
	}
	return nil
}

func (r *ReviewLogRepository) DistinctMergeRequestAuthors(ctx context.Context, query service.ReviewLogOptionQuery) ([]service.AuthorOption, error) {
	var rows []authorRow
	db := applyOptionFilters(r.db.WithContext(ctx).Model(&model.MergeRequestReviewLog{}), query)
	if err := db.Select("author_identity, author_display_name").Where("author_identity <> ''").Order("author_identity ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mergeAuthorOptions(rows), nil
}

func (r *ReviewLogRepository) DistinctMergeRequestProjectNames(ctx context.Context, query service.ReviewLogOptionQuery) ([]string, error) {
	var names []string
	db := applyOptionFilters(r.db.WithContext(ctx).Model(&model.MergeRequestReviewLog{}), query)
	if err := db.Distinct("project_name").Where("project_name <> ''").Order("project_name ASC").Pluck("project_name", &names).Error; err != nil {
		return nil, err
	}
	return names, nil
}

func (r *ReviewLogRepository) UpdateMergeRequestShareToken(ctx context.Context, id uint, token string, expiresAt int64) (*service.MergeRequestReviewLog, error) {
	result := r.db.WithContext(ctx).Model(&model.MergeRequestReviewLog{}).Where("id = ?", id).Updates(map[string]any{
		"share_token":            token,
		"share_token_expires_at": expiresAt,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, service.ErrReviewLogNotFound
	}
	return r.FindMergeRequestByID(ctx, id)
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

func applyOptionFilters(db *gorm.DB, query service.ReviewLogOptionQuery) *gorm.DB {
	if len(query.Authors) > 0 {
		db = db.Where("author_identity IN ?", query.Authors)
	}
	if len(query.ProjectNames) > 0 {
		db = db.Where("project_name IN ?", query.ProjectNames)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}
	return db
}

type authorRow struct {
	AuthorIdentity    string `gorm:"column:author_identity"`
	AuthorDisplayName string `gorm:"column:author_display_name"`
}

func mergeAuthorOptions(rows []authorRow) []service.AuthorOption {
	displayByIdentity := map[string]string{}
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		identity := strings.TrimSpace(row.AuthorIdentity)
		if identity == "" {
			continue
		}
		if _, ok := displayByIdentity[identity]; !ok {
			order = append(order, identity)
		}
		displayByIdentity[identity] = preferDisplayName(displayByIdentity[identity], row.AuthorDisplayName)
	}
	options := make([]service.AuthorOption, 0, len(order))
	for _, identity := range order {
		displayName := displayByIdentity[identity]
		options = append(options, service.AuthorOption{
			Value:       identity,
			Label:       authorDisplayText(identity, displayName),
			DisplayName: displayName,
		})
	}
	return options
}

func preferDisplayName(current string, candidate string) string {
	current = strings.TrimSpace(current)
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return current
	}
	if current == "" || len(candidate) > len(current) {
		return candidate
	}
	return current
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

func analysisExecutionLogModelToService(record *model.ProjectAnalysisPlanExecutionLog) *service.ProjectAnalysisPlanExecutionLog {
	return &service.ProjectAnalysisPlanExecutionLog{
		ID:                  record.ID,
		PlanID:              record.PlanID,
		ProjectID:           record.ProjectID,
		Status:              record.Status,
		ResultContent:       record.ResultContent,
		ResultActions:       record.ResultActions,
		ShareToken:          record.ShareToken,
		ShareTokenExpiresAt: record.ShareTokenExpiresAt,
		ErrorMessage:        record.ErrorMessage,
		ErrorStack:          record.ErrorStack,
		StartedAt:           record.StartedAt,
		CompletedAt:         record.CompletedAt,
		DurationMs:          record.DurationMs,
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
