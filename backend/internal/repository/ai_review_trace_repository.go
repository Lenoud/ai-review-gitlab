package repository

import (
	"context"
	"errors"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

type AIReviewTraceRepository struct {
	db *gorm.DB
}

func NewAIReviewTraceRepository(db *gorm.DB) *AIReviewTraceRepository {
	return &AIReviewTraceRepository{db: db}
}

func (r *AIReviewTraceRepository) Create(ctx context.Context, input service.AIReviewTraceInput) (*service.AIReviewTrace, error) {
	record := &model.AIReviewTrace{
		ReviewEventType: input.ReviewEventType,
		ReviewEventID:   input.ReviewEventID,
		Prompt:          input.Prompt,
		Response:        input.Response,
		Provider:        input.Provider,
		ModelCode:       input.ModelCode,
	}
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return aiReviewTraceModelToService(record), nil
}

func (r *AIReviewTraceRepository) FindByReviewEvent(ctx context.Context, eventType string, eventID uint) (*service.AIReviewTrace, error) {
	var record model.AIReviewTrace
	err := r.db.WithContext(ctx).
		Where("review_event_type = ? AND review_event_id = ?", eventType, eventID).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrAIReviewTraceNotFound
		}
		return nil, err
	}
	return aiReviewTraceModelToService(&record), nil
}

func aiReviewTraceModelToService(record *model.AIReviewTrace) *service.AIReviewTrace {
	return &service.AIReviewTrace{
		ID:              record.ID,
		ReviewEventType: record.ReviewEventType,
		ReviewEventID:   record.ReviewEventID,
		Prompt:          record.Prompt,
		Response:        record.Response,
		Provider:        record.Provider,
		ModelCode:       record.ModelCode,
		CreatedAt:       record.CreatedAt.UnixMilli(),
		UpdatedAt:       record.UpdatedAt.UnixMilli(),
	}
}
