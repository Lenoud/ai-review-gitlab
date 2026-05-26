package repository

import (
	"context"
	"errors"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

type IMRobotRepository struct {
	db *gorm.DB
}

func NewIMRobotRepository(db *gorm.DB) *IMRobotRepository {
	return &IMRobotRepository{db: db}
}

func (r *IMRobotRepository) CreateIMRobot(ctx context.Context, input service.IMRobotInput) (*service.IMRobot, error) {
	record := imRobotModelFromInput(input)
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	if !input.Enabled {
		if err := r.db.WithContext(ctx).Model(&model.IMRobot{}).Where("id = ?", record.ID).Update("enabled", false).Error; err != nil {
			return nil, err
		}
	}
	return r.FindIMRobotByID(ctx, record.ID)
}

func (r *IMRobotRepository) UpdateIMRobot(ctx context.Context, id uint, input service.IMRobotInput) (*service.IMRobot, error) {
	record := imRobotModelFromInput(input)
	result := r.db.WithContext(ctx).Model(&model.IMRobot{}).Where("id = ?", id).Updates(map[string]any{
		"platform":    record.Platform,
		"name":        record.Name,
		"webhook_url": record.WebhookURL,
		"secret":      record.Secret,
		"enabled":     record.Enabled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, service.ErrIMRobotNotFound
	}
	return r.FindIMRobotByID(ctx, id)
}

func (r *IMRobotRepository) FindIMRobotByID(ctx context.Context, id uint) (*service.IMRobot, error) {
	var record model.IMRobot
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrIMRobotNotFound
		}
		return nil, err
	}
	return imRobotModelToService(&record), nil
}

func (r *IMRobotRepository) DeleteIMRobots(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&model.IMRobot{}, ids).Error
}

func (r *IMRobotRepository) CountIMRobotReferences(ctx context.Context, ids []uint) (int64, error) {
	var projectCount int64
	if err := r.db.WithContext(ctx).Model(&model.Project{}).Where("im_robot_id IN ?", ids).Count(&projectCount).Error; err != nil {
		return 0, err
	}
	var analysisPlanCount int64
	if err := r.db.WithContext(ctx).Model(&model.ProjectAnalysisPlan{}).Where("im_robot_id IN ?", ids).Count(&analysisPlanCount).Error; err != nil {
		return 0, err
	}
	return projectCount + analysisPlanCount, nil
}

func (r *IMRobotRepository) SearchIMRobots(ctx context.Context, query service.IMRobotSearchQuery) (*service.IMRobotPage, error) {
	db := r.db.WithContext(ctx).Model(&model.IMRobot{})
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR webhook_url LIKE ?", like, like)
	}
	if query.Platform != "" {
		db = db.Where("platform = ?", query.Platform)
	}
	if query.Enabled != nil {
		db = db.Where("enabled = ?", *query.Enabled)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.IMRobot
	offset := (query.Page - 1) * query.Size
	if err := db.Order("id DESC").Limit(query.Size).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.IMRobot, 0, len(records))
	for i := range records {
		items = append(items, *imRobotModelToService(&records[i]))
	}
	return &service.IMRobotPage{Items: items, Total: total, Page: query.Page, Size: query.Size}, nil
}

func (r *IMRobotRepository) ListEnabledIMRobots(ctx context.Context) ([]service.IMRobot, error) {
	var records []model.IMRobot
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Order("id DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]service.IMRobot, 0, len(records))
	for i := range records {
		items = append(items, *imRobotModelToService(&records[i]))
	}
	return items, nil
}

func imRobotModelFromInput(input service.IMRobotInput) *model.IMRobot {
	return &model.IMRobot{
		Platform:   input.Platform,
		Name:       input.Name,
		WebhookURL: input.WebhookURL,
		Secret:     input.Secret,
		Enabled:    input.Enabled,
	}
}

func imRobotModelToService(record *model.IMRobot) *service.IMRobot {
	return &service.IMRobot{
		ID:         record.ID,
		Platform:   record.Platform,
		Name:       record.Name,
		WebhookURL: record.WebhookURL,
		Secret:     record.Secret,
		Enabled:    record.Enabled,
		CreatedAt:  record.CreatedAt.UnixMilli(),
		UpdatedAt:  record.UpdatedAt.UnixMilli(),
	}
}
