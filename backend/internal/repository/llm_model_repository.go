package repository

import (
	"context"
	"errors"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

type LLMModelRepository struct {
	db *gorm.DB
}

func NewLLMModelRepository(db *gorm.DB) *LLMModelRepository {
	return &LLMModelRepository{db: db}
}

func (r *LLMModelRepository) Create(ctx context.Context, input service.LLMModelInput) (*service.LLMModel, error) {
	var created model.LLMModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if input.IsDefault {
			if err := clearDefaultLLMModel(tx).Error; err != nil {
				return err
			}
		}
		created = llmModelFromInput(input)
		return tx.Create(&created).Error
	})
	if err != nil {
		return nil, err
	}
	return llmModelToService(&created), nil
}

func (r *LLMModelRepository) Update(ctx context.Context, id uint, input service.LLMModelInput) (*service.LLMModel, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if input.IsDefault {
			if err := clearDefaultLLMModel(tx).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.LLMModel{}).Where("id = ?", id).Updates(map[string]any{
			"provider":     input.Provider,
			"model_code":   input.ModelCode,
			"api_base_url": input.APIBaseURL,
			"api_key":      input.APIKey,
			"max_tokens":   input.MaxTokens,
			"is_default":   input.IsDefault,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *LLMModelRepository) FindByID(ctx context.Context, id uint) (*service.LLMModel, error) {
	var record model.LLMModel
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrLLMModelNotFound
		}
		return nil, err
	}
	return llmModelToService(&record), nil
}

func (r *LLMModelRepository) Delete(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&model.LLMModel{}, ids).Error
}

func (r *LLMModelRepository) Search(ctx context.Context, query service.LLMModelSearchQuery) (*service.LLMModelPage, error) {
	db := r.db.WithContext(ctx).Model(&model.LLMModel{})
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("provider LIKE ? OR model_code LIKE ? OR api_base_url LIKE ?", like, like, like)
	}
	if query.Provider != "" {
		db = db.Where("provider = ?", query.Provider)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.LLMModel
	offset := (query.Page - 1) * query.Size
	if err := db.Order("is_default DESC, id DESC").Offset(offset).Limit(query.Size).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.LLMModel, 0, len(records))
	for i := range records {
		items = append(items, *llmModelToService(&records[i]))
	}
	return &service.LLMModelPage{
		Items: items,
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
	}, nil
}

func (r *LLMModelRepository) Default(ctx context.Context) (*service.LLMModel, error) {
	var record model.LLMModel
	err := r.db.WithContext(ctx).Where("is_default = ?", true).Order("id DESC").First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrLLMModelNotFound
		}
		return nil, err
	}
	return llmModelToService(&record), nil
}

func (r *LLMModelRepository) SetDefault(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record model.LLMModel
		if err := tx.First(&record, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return service.ErrLLMModelNotFound
			}
			return err
		}
		if err := clearDefaultLLMModel(tx).Error; err != nil {
			return err
		}
		return tx.Model(&model.LLMModel{}).Where("id = ?", id).Update("is_default", true).Error
	})
}

func clearDefaultLLMModel(db *gorm.DB) *gorm.DB {
	return db.Model(&model.LLMModel{}).Where("is_default = ?", true).Update("is_default", false)
}

func llmModelFromInput(input service.LLMModelInput) model.LLMModel {
	return model.LLMModel{
		Provider:   input.Provider,
		ModelCode:  input.ModelCode,
		APIBaseURL: input.APIBaseURL,
		APIKey:     input.APIKey,
		MaxTokens:  input.MaxTokens,
		IsDefault:  input.IsDefault,
	}
}

func llmModelToService(record *model.LLMModel) *service.LLMModel {
	return &service.LLMModel{
		ID:         record.ID,
		Provider:   record.Provider,
		ModelCode:  record.ModelCode,
		APIBaseURL: record.APIBaseURL,
		APIKey:     record.APIKey,
		MaxTokens:  record.MaxTokens,
		IsDefault:  record.IsDefault,
	}
}
