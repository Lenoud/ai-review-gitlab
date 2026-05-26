package repository

import (
	"context"
	"errors"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

type MemberIMMappingRepository struct {
	db *gorm.DB
}

func NewMemberIMMappingRepository(db *gorm.DB) *MemberIMMappingRepository {
	return &MemberIMMappingRepository{db: db}
}

func (r *MemberIMMappingRepository) CreateMemberIMMapping(ctx context.Context, input service.MemberIMMappingInput) (*service.MemberIMMapping, error) {
	record := memberIMMappingModelFromInput(input)
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return r.FindMemberIMMappingByID(ctx, record.ID)
}

func (r *MemberIMMappingRepository) UpdateMemberIMMapping(ctx context.Context, id uint, input service.MemberIMMappingInput) (*service.MemberIMMapping, error) {
	record := memberIMMappingModelFromInput(input)
	result := r.db.WithContext(ctx).Model(&model.MemberIMMapping{}).Where("id = ?", id).Updates(map[string]any{
		"git_username": record.GitUsername,
		"platform":     record.Platform,
		"im_user_id":   record.IMUserID,
		"display_name": record.DisplayName,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return r.FindMemberIMMappingByID(ctx, id)
}

func (r *MemberIMMappingRepository) FindMemberIMMappingByID(ctx context.Context, id uint) (*service.MemberIMMapping, error) {
	var record model.MemberIMMapping
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrMemberIMMappingNotFound
		}
		return nil, err
	}
	return memberIMMappingModelToService(&record), nil
}

func (r *MemberIMMappingRepository) DeleteMemberIMMappings(ctx context.Context, ids []uint) error {
	return r.db.WithContext(ctx).Delete(&model.MemberIMMapping{}, ids).Error
}

func (r *MemberIMMappingRepository) SearchMemberIMMappings(ctx context.Context, query service.MemberIMMappingSearchQuery) (*service.MemberIMMappingPage, error) {
	db := r.db.WithContext(ctx).Model(&model.MemberIMMapping{})
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		db = db.Where("git_username LIKE ? OR im_user_id LIKE ? OR display_name LIKE ?", like, like, like)
	}
	if query.Platform != "" {
		db = db.Where("platform = ?", query.Platform)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.MemberIMMapping
	offset := (query.Page - 1) * query.Size
	if err := db.Order("id DESC").Limit(query.Size).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	items := make([]service.MemberIMMapping, 0, len(records))
	for i := range records {
		items = append(items, *memberIMMappingModelToService(&records[i]))
	}
	return &service.MemberIMMappingPage{Items: items, Total: total, Page: query.Page, Size: query.Size}, nil
}

func (r *MemberIMMappingRepository) ExistsMemberIMMapping(ctx context.Context, gitUsername string, platform string, excludeID uint) (bool, error) {
	db := r.db.WithContext(ctx).Model(&model.MemberIMMapping{}).Where("git_username = ? AND platform = ?", gitUsername, platform)
	if excludeID > 0 {
		db = db.Where("id <> ?", excludeID)
	}
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func memberIMMappingModelFromInput(input service.MemberIMMappingInput) *model.MemberIMMapping {
	return &model.MemberIMMapping{
		GitUsername: input.GitUsername,
		Platform:    input.Platform,
		IMUserID:    input.IMUserID,
		DisplayName: input.DisplayName,
	}
}

func memberIMMappingModelToService(record *model.MemberIMMapping) *service.MemberIMMapping {
	return &service.MemberIMMapping{
		ID:          record.ID,
		GitUsername: record.GitUsername,
		Platform:    record.Platform,
		IMUserID:    record.IMUserID,
		DisplayName: record.DisplayName,
		CreatedAt:   record.CreatedAt.UnixMilli(),
		UpdatedAt:   record.UpdatedAt.UnixMilli(),
	}
}
