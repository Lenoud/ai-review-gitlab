package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ReviewTaskRepository struct {
	db *gorm.DB
}

func NewReviewTaskRepository(db *gorm.DB) *ReviewTaskRepository {
	return &ReviewTaskRepository{db: db}
}

func (r *ReviewTaskRepository) CreateOrGetByDedupeKey(ctx context.Context, input service.ReviewTaskCreateInput) (*service.ReviewTask, bool, error) {
	var out service.ReviewTask
	var duplicate bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.ReviewTask
		err := tx.Where("dedupe_key = ?", input.DedupeKey).First(&existing).Error
		if err == nil {
			task, err := reviewTaskModelToService(&existing)
			if err != nil {
				return err
			}
			out = *task
			duplicate = true
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		record := reviewTaskModelFromCreateInput(input)
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		task, err := reviewTaskModelToService(&record)
		if err != nil {
			return err
		}
		out = *task
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return &out, duplicate, nil
}

func (r *ReviewTaskRepository) ClaimNext(ctx context.Context, workerID string, now time.Time) (*service.ReviewTask, error) {
	var out service.ReviewTask
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record model.ReviewTask
		err := tx.Where("status = ? AND next_run_at <= ?", service.ReviewTaskStatusPending, now).
			Order("priority DESC, id ASC").
			First(&record).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return service.ErrReviewTaskNotFound
			}
			return err
		}
		updates := map[string]any{
			"status":     service.ReviewTaskStatusRunning,
			"locked_by":  workerID,
			"locked_at":  now,
			"started_at": now,
		}
		if err := tx.Model(&model.ReviewTask{}).Where("id = ?", record.ID).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.First(&record, record.ID).Error; err != nil {
			return err
		}
		task, err := reviewTaskModelToService(&record)
		if err != nil {
			return err
		}
		out = *task
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *ReviewTaskRepository) StartAttempt(ctx context.Context, taskID uint, now time.Time) (*service.ReviewTaskAttempt, error) {
	task, err := r.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	record := model.ReviewTaskAttempt{
		TaskID:    taskID,
		AttemptNo: task.Attempts + 1,
		Status:    service.ReviewTaskAttemptStatusRunning,
		StartedAt: now,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, err
	}
	return reviewTaskAttemptModelToService(&record), nil
}

func (r *ReviewTaskRepository) MarkSucceeded(ctx context.Context, taskID uint, finishedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.ReviewTask{}).Where("id = ?", taskID).Updates(map[string]any{
		"status":      service.ReviewTaskStatusSucceeded,
		"finished_at": finishedAt,
		"locked_by":   "",
		"locked_at":   nil,
	}).Error
}

func (r *ReviewTaskRepository) MarkFailed(ctx context.Context, taskID uint, attempts int, status string, nextRunAt *time.Time, errorMessage string, finishedAt *time.Time) error {
	updates := map[string]any{
		"attempts":      attempts,
		"status":        status,
		"next_run_at":   nextRunAt,
		"error_message": errorMessage,
		"finished_at":   finishedAt,
		"locked_by":     "",
		"locked_at":     nil,
	}
	if nextRunAt == nil {
		updates["next_run_at"] = time.Time{}
	}
	return r.db.WithContext(ctx).Model(&model.ReviewTask{}).Where("id = ?", taskID).Updates(updates).Error
}

func (r *ReviewTaskRepository) FindByID(ctx context.Context, id uint) (*service.ReviewTask, error) {
	var record model.ReviewTask
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrReviewTaskNotFound
		}
		return nil, err
	}
	return reviewTaskModelToService(&record)
}

func reviewTaskModelFromCreateInput(input service.ReviewTaskCreateInput) model.ReviewTask {
	maxAttempts := input.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	return model.ReviewTask{
		ProjectID:   input.ProjectID,
		EventType:   input.EventType,
		DedupeKey:   input.DedupeKey,
		PayloadJSON: datatypes.JSON(append([]byte(nil), input.PayloadJSON...)),
		Status:      service.ReviewTaskStatusPending,
		Priority:    input.Priority,
		MaxAttempts: maxAttempts,
		NextRunAt:   input.NextRunAt,
	}
}

func reviewTaskModelToService(record *model.ReviewTask) (*service.ReviewTask, error) {
	return &service.ReviewTask{
		ID:            record.ID,
		ProjectID:     record.ProjectID,
		EventType:     record.EventType,
		DedupeKey:     record.DedupeKey,
		PayloadJSON:   append([]byte(nil), record.PayloadJSON...),
		Status:        record.Status,
		Priority:      record.Priority,
		Attempts:      record.Attempts,
		MaxAttempts:   record.MaxAttempts,
		NextRunAt:     record.NextRunAt,
		LockedBy:      record.LockedBy,
		LockedAt:      record.LockedAt,
		StartedAt:     record.StartedAt,
		FinishedAt:    record.FinishedAt,
		ErrorMessage:  record.ErrorMessage,
		ResultLogType: record.ResultLogType,
		ResultLogID:   record.ResultLogID,
	}, nil
}

func reviewTaskAttemptModelToService(record *model.ReviewTaskAttempt) *service.ReviewTaskAttempt {
	return &service.ReviewTaskAttempt{
		ID:           record.ID,
		TaskID:       record.TaskID,
		AttemptNo:    record.AttemptNo,
		Status:       record.Status,
		StartedAt:    record.StartedAt,
		FinishedAt:   record.FinishedAt,
		DurationMs:   record.DurationMs,
		ErrorMessage: record.ErrorMessage,
		ErrorStack:   record.ErrorStack,
	}
}
