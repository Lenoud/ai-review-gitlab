package repository

import (
	"context"
	"errors"

	"github.com/Lenoud/ai-review-gitlab/backend/internal/model"
	"github.com/Lenoud/ai-review-gitlab/backend/internal/service"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*service.User, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}
	return toServiceUser(&user), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*service.User, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, service.ErrUserNotFound
		}
		return nil, err
	}
	return toServiceUser(&user), nil
}

func toServiceUser(user *model.SysUser) *service.User {
	return &service.User{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Nickname:     user.Nickname,
		Status:       user.Status,
	}
}
