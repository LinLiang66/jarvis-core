package store

import (
	"context"

	"gorm.io/gorm"

	"jarvis-core/backend/internal/infra/base"
	"jarvis-core/backend/internal/model"
)

type UserRepository struct {
	base.CRUD
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{CRUD: base.CRUD{DB: db}}
}

func (r *UserRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.User{})
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var row model.User
	if err := r.DB.WithContext(ctx).Where("username = ?", username).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id any) (*model.User, error) {
	var row model.User
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&model.User{}).Count(&n).Error
	return n, err
}

func (r *UserRepository) Create(ctx context.Context, row *model.User) error {
	return r.DB.WithContext(ctx).Create(row).Error
}
