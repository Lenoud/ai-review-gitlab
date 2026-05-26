package repository

import (
	"context"
	"errors"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

type ProjectAnalysisPlanRepository struct {
	db *gorm.DB
}

func NewProjectAnalysisPlanRepository(db *gorm.DB) *ProjectAnalysisPlanRepository {
	return &ProjectAnalysisPlanRepository{db: db}
}

func (r *ProjectAnalysisPlanRepository) CreateProjectAnalysisPlan(ctx context.Context, input service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error) {
	record := projectAnalysisPlanModelFromInput(input)
	if err := r.db.WithContext(ctx).Select(
		"ProjectID",
		"Name",
		"Prompt",
		"CronExpression",
		"Enabled",
		"IMEnabled",
		"IMRobotID",
		"HTMLReportEnabled",
	).Create(record).Error; err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if input.Enabled != nil {
		updates["enabled"] = *input.Enabled
	}
	if input.HTMLReportEnabled != nil {
		updates["html_report_enabled"] = *input.HTMLReportEnabled
	}
	if len(updates) > 0 {
		if err := r.db.WithContext(ctx).Model(&model.ProjectAnalysisPlan{}).Where("id = ?", record.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return r.FindProjectAnalysisPlanByID(ctx, record.ID)
}

func (r *ProjectAnalysisPlanRepository) UpdateProjectAnalysisPlan(ctx context.Context, id uint, input service.ProjectAnalysisPlanInput) (*service.ProjectAnalysisPlan, error) {
	record := projectAnalysisPlanModelFromInput(input)
	result := r.db.WithContext(ctx).Model(&model.ProjectAnalysisPlan{}).Where("id = ?", id).Updates(map[string]any{
		"project_id":          record.ProjectID,
		"name":                record.Name,
		"prompt":              record.Prompt,
		"cron_expression":     record.CronExpression,
		"enabled":             record.Enabled,
		"im_enabled":          record.IMEnabled,
		"im_robot_id":         record.IMRobotID,
		"html_report_enabled": record.HTMLReportEnabled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, service.ErrProjectAnalysisPlanNotFound
	}
	return r.FindProjectAnalysisPlanByID(ctx, id)
}

func (r *ProjectAnalysisPlanRepository) FindProjectAnalysisPlanByID(ctx context.Context, id uint) (*service.ProjectAnalysisPlan, error) {
	var record model.ProjectAnalysisPlan
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrProjectAnalysisPlanNotFound
		}
		return nil, err
	}
	return projectAnalysisPlanModelToService(&record), nil
}

func (r *ProjectAnalysisPlanRepository) DeleteProjectAnalysisPlans(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&model.ProjectAnalysisPlan{}, ids).Error
}

func (r *ProjectAnalysisPlanRepository) SearchProjectAnalysisPlans(ctx context.Context, query service.ProjectAnalysisPlanSearchQuery) (*service.ProjectAnalysisPlanPage, error) {
	db := r.db.WithContext(ctx).Model(&model.ProjectAnalysisPlan{})
	if query.ProjectID > 0 {
		db = db.Where("project_id = ?", query.ProjectID)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR prompt LIKE ?", like, like)
	}
	if query.Enabled != nil {
		db = db.Where("enabled = ?", *query.Enabled)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.ProjectAnalysisPlan
	offset := (query.Page - 1) * query.Size
	if err := db.Order("id DESC").Offset(offset).Limit(query.Size).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.ProjectAnalysisPlan, 0, len(records))
	for i := range records {
		items = append(items, *projectAnalysisPlanModelToService(&records[i]))
	}
	return &service.ProjectAnalysisPlanPage{Items: items, Total: total, Page: query.Page, Size: query.Size}, nil
}

func projectAnalysisPlanModelFromInput(input service.ProjectAnalysisPlanInput) *model.ProjectAnalysisPlan {
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	htmlReportEnabled := true
	if input.HTMLReportEnabled != nil {
		htmlReportEnabled = *input.HTMLReportEnabled
	}
	return &model.ProjectAnalysisPlan{
		ProjectID:         input.ProjectID,
		Name:              input.Name,
		Prompt:            input.Prompt,
		CronExpression:    input.CronExpression,
		Enabled:           enabled,
		IMEnabled:         input.IMEnabled,
		IMRobotID:         input.IMRobotID,
		HTMLReportEnabled: htmlReportEnabled,
	}
}

func projectAnalysisPlanModelToService(record *model.ProjectAnalysisPlan) *service.ProjectAnalysisPlan {
	return &service.ProjectAnalysisPlan{
		ID:                record.ID,
		ProjectID:         record.ProjectID,
		Name:              record.Name,
		Prompt:            record.Prompt,
		CronExpression:    record.CronExpression,
		Enabled:           record.Enabled,
		IMEnabled:         record.IMEnabled,
		IMRobotID:         record.IMRobotID,
		HTMLReportEnabled: record.HTMLReportEnabled,
		CreatedAt:         record.CreatedAt.UnixMilli(),
		UpdatedAt:         record.UpdatedAt.UnixMilli(),
	}
}
