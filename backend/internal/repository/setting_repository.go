package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"gorm.io/gorm"
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) GetSettingValue(ctx context.Context, key string) (string, bool, error) {
	var record model.Setting
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	return record.Value, true, nil
}

func (r *SettingRepository) SetSettingValue(ctx context.Context, key string, value string) error {
	var record model.Setting
	err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.setSettingValueAfterMissing(ctx, key, value)
		}
		return err
	}
	return r.db.WithContext(ctx).Model(&record).Update("value", value).Error
}

func (r *SettingRepository) setSettingValueAfterMissing(ctx context.Context, key string, value string) error {
	err := r.db.WithContext(ctx).Create(&model.Setting{Key: key, Value: value}).Error
	if err == nil {
		return nil
	}
	if !isDuplicateKeyError(err) {
		return err
	}
	return r.db.WithContext(ctx).Model(&model.Setting{}).Where("`key` = ?", key).Update("value", value).Error
}

func isDuplicateKeyError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique constraint")
}
