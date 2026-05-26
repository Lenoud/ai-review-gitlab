package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ProjectTemplateReviewRuleRepository struct {
	db *gorm.DB
}

func NewProjectTemplateReviewRuleRepository(db *gorm.DB) *ProjectTemplateReviewRuleRepository {
	return &ProjectTemplateReviewRuleRepository{db: db}
}

func (r *ProjectTemplateReviewRuleRepository) CreateProjectTemplateReviewRule(ctx context.Context, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error) {
	record, err := projectTemplateReviewRuleModelFromInput(input)
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).
		Select("TemplateID", "Name", "Description", "GlobPatterns", "Content", "Priority", "Enabled").
		Create(record).Error; err != nil {
		return nil, err
	}
	if !input.Enabled {
		if err := r.db.WithContext(ctx).Model(&model.ProjectTemplateReviewRule{}).Where("id = ?", record.ID).Update("enabled", false).Error; err != nil {
			return nil, err
		}
	}
	return r.FindProjectTemplateReviewRuleByID(ctx, record.ID)
}

func (r *ProjectTemplateReviewRuleRepository) UpdateProjectTemplateReviewRule(ctx context.Context, id uint, input service.ProjectTemplateReviewRuleInput) (*service.ProjectTemplateReviewRule, error) {
	record, err := projectTemplateReviewRuleModelFromInput(input)
	if err != nil {
		return nil, err
	}
	result := r.db.WithContext(ctx).Model(&model.ProjectTemplateReviewRule{}).Where("id = ?", id).Updates(map[string]any{
		"template_id":   record.TemplateID,
		"name":          record.Name,
		"description":   record.Description,
		"glob_patterns": record.GlobPatterns,
		"content":       record.Content,
		"priority":      record.Priority,
		"enabled":       record.Enabled,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, service.ErrProjectTemplateReviewRuleNotFound
	}
	return r.FindProjectTemplateReviewRuleByID(ctx, id)
}

func (r *ProjectTemplateReviewRuleRepository) FindProjectTemplateReviewRuleByID(ctx context.Context, id uint) (*service.ProjectTemplateReviewRule, error) {
	var record model.ProjectTemplateReviewRule
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrProjectTemplateReviewRuleNotFound
		}
		return nil, err
	}
	return projectTemplateReviewRuleModelToService(&record)
}

func (r *ProjectTemplateReviewRuleRepository) DeleteProjectTemplateReviewRule(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.ProjectTemplateReviewRule{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrProjectTemplateReviewRuleNotFound
	}
	return nil
}

func (r *ProjectTemplateReviewRuleRepository) ListProjectTemplateReviewRulesByTemplateID(ctx context.Context, templateID uint) ([]service.ProjectTemplateReviewRule, error) {
	var records []model.ProjectTemplateReviewRule
	err := r.db.WithContext(ctx).
		Where("template_id = ?", templateID).
		Order("priority DESC").
		Order("created_at ASC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	items := make([]service.ProjectTemplateReviewRule, 0, len(records))
	for i := range records {
		item, err := projectTemplateReviewRuleModelToService(&records[i])
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func (r *ProjectTemplateReviewRuleRepository) ProjectTemplateExists(ctx context.Context, templateID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ProjectTemplate{}).Where("id = ?", templateID).Count(&count).Error
	return count > 0, err
}

func projectTemplateReviewRuleModelFromInput(input service.ProjectTemplateReviewRuleInput) (*model.ProjectTemplateReviewRule, error) {
	globPatterns, err := json.Marshal(input.GlobPatterns)
	if err != nil {
		return nil, err
	}
	return &model.ProjectTemplateReviewRule{
		TemplateID:   input.TemplateID,
		Name:         input.Name,
		Description:  input.Description,
		GlobPatterns: datatypes.JSON(globPatterns),
		Content:      input.Content,
		Priority:     input.Priority,
		Enabled:      input.Enabled,
	}, nil
}

func projectTemplateReviewRuleModelToService(record *model.ProjectTemplateReviewRule) (*service.ProjectTemplateReviewRule, error) {
	var globPatterns []string
	if len(record.GlobPatterns) > 0 {
		if err := json.Unmarshal(record.GlobPatterns, &globPatterns); err != nil {
			return nil, err
		}
	}
	return &service.ProjectTemplateReviewRule{
		ID:           record.ID,
		TemplateID:   record.TemplateID,
		Name:         record.Name,
		Description:  record.Description,
		GlobPatterns: globPatterns,
		Content:      record.Content,
		Priority:     record.Priority,
		Enabled:      record.Enabled,
		CreatedAt:    record.CreatedAt.UnixMilli(),
		UpdatedAt:    record.UpdatedAt.UnixMilli(),
	}, nil
}
