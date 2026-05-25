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

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, input service.ProjectInput) (*service.Project, error) {
	record, err := projectModelFromInput(input)
	if err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return projectModelToService(record)
}

func (r *ProjectRepository) BatchCreate(ctx context.Context, inputs []service.ProjectInput) ([]service.Project, error) {
	records := make([]model.Project, 0, len(inputs))
	for _, input := range inputs {
		record, err := projectModelFromInput(input)
		if err != nil {
			return nil, err
		}
		records = append(records, *record)
	}
	if err := r.db.WithContext(ctx).Create(&records).Error; err != nil {
		return nil, err
	}
	projects := make([]service.Project, 0, len(records))
	for i := range records {
		project, err := projectModelToService(&records[i])
		if err != nil {
			return nil, err
		}
		projects = append(projects, *project)
	}
	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, id uint, input service.ProjectInput) (*service.Project, error) {
	record, err := projectModelFromInput(input)
	if err != nil {
		return nil, err
	}
	record.ID = id
	err = r.db.WithContext(ctx).Model(&model.Project{}).Where("id = ?", id).Updates(map[string]any{
		"name":                         record.Name,
		"description":                  record.Description,
		"web_url":                      record.WebURL,
		"platform":                     record.Platform,
		"access_token":                 record.AccessToken,
		"im_enabled":                   record.IMEnabled,
		"im_robot_id":                  record.IMRobotID,
		"im_at_member_enabled":         record.IMAtMemberEnabled,
		"im_at_member_score_threshold": record.IMAtMemberScoreThreshold,
		"ai_review_enabled":            record.AIReviewEnabled,
		"template_id":                  record.TemplateID,
		"extensions":                   record.Extensions,
		"review_event_types":           record.ReviewEventTypes,
		"review_prompt_template":       record.ReviewPromptTemplate,
		"html_report_enabled":          record.HTMLReportEnabled,
		"deep_review_enabled":          record.DeepReviewEnabled,
	}).Error
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *ProjectRepository) FindByID(ctx context.Context, id uint) (*service.Project, error) {
	var record model.Project
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrProjectNotFound
		}
		return nil, err
	}
	return projectModelToService(&record)
}

func (r *ProjectRepository) Delete(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, ids).Error
}

func (r *ProjectRepository) Search(ctx context.Context, query service.ProjectSearchQuery) (*service.ProjectPage, error) {
	db := r.db.WithContext(ctx).Model(&model.Project{})
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR web_url LIKE ? OR description LIKE ?", like, like, like)
	}
	if query.Platform != "" {
		db = db.Where("platform = ?", query.Platform)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.Project
	offset := (query.Page - 1) * query.Size
	if err := db.Order("id DESC").Offset(offset).Limit(query.Size).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.Project, 0, len(records))
	for i := range records {
		project, err := projectModelToService(&records[i])
		if err != nil {
			return nil, err
		}
		items = append(items, *project)
	}
	return &service.ProjectPage{
		Items: items,
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
	}, nil
}

func (r *ProjectRepository) ExistsByWebURL(ctx context.Context, webURL string, excludeID uint) (bool, error) {
	db := r.db.WithContext(ctx).Model(&model.Project{}).Where("web_url = ?", webURL)
	if excludeID > 0 {
		db = db.Where("id <> ?", excludeID)
	}
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func projectModelFromInput(input service.ProjectInput) (*model.Project, error) {
	extensions, err := json.Marshal(input.Extensions)
	if err != nil {
		return nil, err
	}
	reviewEventTypes, err := json.Marshal(input.ReviewEventTypes)
	if err != nil {
		return nil, err
	}
	aiReviewEnabled := true
	if input.AIReviewEnabled != nil {
		aiReviewEnabled = *input.AIReviewEnabled
	}
	return &model.Project{
		Name:                     input.Name,
		Description:              input.Description,
		WebURL:                   input.WebURL,
		Platform:                 input.Platform,
		AccessToken:              input.AccessToken,
		IMEnabled:                input.IMEnabled,
		IMRobotID:                input.IMRobotID,
		IMAtMemberEnabled:        input.IMAtMemberEnabled,
		IMAtMemberScoreThreshold: input.IMAtMemberScoreThreshold,
		AIReviewEnabled:          aiReviewEnabled,
		TemplateID:               input.TemplateID,
		Extensions:               datatypes.JSON(extensions),
		ReviewEventTypes:         datatypes.JSON(reviewEventTypes),
		ReviewPromptTemplate:     input.ReviewPromptTemplate,
		HTMLReportEnabled:        input.HTMLReportEnabled,
		DeepReviewEnabled:        input.DeepReviewEnabled,
	}, nil
}

func projectModelToService(record *model.Project) (*service.Project, error) {
	var extensions []string
	if len(record.Extensions) > 0 {
		if err := json.Unmarshal(record.Extensions, &extensions); err != nil {
			return nil, err
		}
	}
	var reviewEventTypes []string
	if len(record.ReviewEventTypes) > 0 {
		if err := json.Unmarshal(record.ReviewEventTypes, &reviewEventTypes); err != nil {
			return nil, err
		}
	}
	return &service.Project{
		ID:                       record.ID,
		Name:                     record.Name,
		Description:              record.Description,
		WebURL:                   record.WebURL,
		Platform:                 record.Platform,
		AccessToken:              record.AccessToken,
		IMEnabled:                record.IMEnabled,
		IMRobotID:                record.IMRobotID,
		IMAtMemberEnabled:        record.IMAtMemberEnabled,
		IMAtMemberScoreThreshold: record.IMAtMemberScoreThreshold,
		AIReviewEnabled:          record.AIReviewEnabled,
		TemplateID:               record.TemplateID,
		Extensions:               extensions,
		ReviewEventTypes:         reviewEventTypes,
		ReviewPromptTemplate:     record.ReviewPromptTemplate,
		HTMLReportEnabled:        record.HTMLReportEnabled,
		DeepReviewEnabled:        record.DeepReviewEnabled,
	}, nil
}
