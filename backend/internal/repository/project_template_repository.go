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

type ProjectTemplateRepository struct {
	db *gorm.DB
}

func NewProjectTemplateRepository(db *gorm.DB) *ProjectTemplateRepository {
	return &ProjectTemplateRepository{db: db}
}

func (r *ProjectTemplateRepository) CreateProjectTemplate(ctx context.Context, input service.ProjectTemplateInput) (*service.ProjectTemplate, error) {
	record, err := projectTemplateModelFromInput(input)
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return r.FindProjectTemplateByID(ctx, record.ID)
}

func (r *ProjectTemplateRepository) UpdateProjectTemplate(ctx context.Context, id uint, input service.ProjectTemplateInput) (*service.ProjectTemplate, error) {
	record, err := projectTemplateModelFromInput(input)
	if err != nil {
		return nil, err
	}
	result := r.db.WithContext(ctx).Model(&model.ProjectTemplate{}).Where("id = ?", id).Updates(map[string]any{
		"name":                   record.Name,
		"description":            record.Description,
		"extensions":             record.Extensions,
		"review_prompt_template": record.ReviewPromptTemplate,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, service.ErrProjectTemplateNotFound
	}
	return r.FindProjectTemplateByID(ctx, id)
}

func (r *ProjectTemplateRepository) FindProjectTemplateByID(ctx context.Context, id uint) (*service.ProjectTemplate, error) {
	var record model.ProjectTemplate
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrProjectTemplateNotFound
		}
		return nil, err
	}
	return projectTemplateModelToService(&record)
}

func (r *ProjectTemplateRepository) DeleteProjectTemplates(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&model.ProjectTemplate{}, ids).Error
}

func (r *ProjectTemplateRepository) ListProjectTemplates(ctx context.Context, query service.ProjectTemplateListQuery) ([]service.ProjectTemplate, error) {
	db := r.db.WithContext(ctx).Model(&model.ProjectTemplate{})
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ? OR review_prompt_template LIKE ?", like, like, like)
	}

	var records []model.ProjectTemplate
	if err := db.Order("id DESC").Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.ProjectTemplate, 0, len(records))
	for i := range records {
		item, err := projectTemplateModelToService(&records[i])
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func (r *ProjectTemplateRepository) CountProjectsUsingTemplates(ctx context.Context, ids []uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Project{}).Where("template_id IN ?", ids).Count(&count).Error
	return count, err
}

func projectTemplateModelFromInput(input service.ProjectTemplateInput) (*model.ProjectTemplate, error) {
	extensions, err := json.Marshal(input.Extensions)
	if err != nil {
		return nil, err
	}
	return &model.ProjectTemplate{
		Name:                 input.Name,
		Description:          input.Description,
		Extensions:           datatypes.JSON(extensions),
		ReviewPromptTemplate: input.ReviewPromptTemplate,
	}, nil
}

func projectTemplateModelToService(record *model.ProjectTemplate) (*service.ProjectTemplate, error) {
	var extensions []string
	if len(record.Extensions) > 0 {
		if err := json.Unmarshal(record.Extensions, &extensions); err != nil {
			return nil, err
		}
	}
	return &service.ProjectTemplate{
		ID:                   record.ID,
		Name:                 record.Name,
		Description:          record.Description,
		Extensions:           extensions,
		ReviewPromptTemplate: record.ReviewPromptTemplate,
		CreatedAt:            record.CreatedAt.UnixMilli(),
		UpdatedAt:            record.UpdatedAt.UnixMilli(),
	}, nil
}
