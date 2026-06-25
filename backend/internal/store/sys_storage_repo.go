package store

import (
	"context"

	"gorm.io/gorm"

	"jarvis-core/backend/internal/infra/base"
	"jarvis-core/backend/internal/model"
)

type SysStorageRepository struct{ base.CRUD }

func NewSysStorageRepository(db *gorm.DB) *SysStorageRepository {
	return &SysStorageRepository{CRUD: base.CRUD{DB: db}}
}

func (r *SysStorageRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.SysStorage{})
}

func (r *SysStorageRepository) List(ctx context.Context, storageType int) ([]model.SysStorage, error) {
	q := r.DB.WithContext(ctx).Model(&model.SysStorage{})
	if storageType > 0 {
		q = q.Where("type = ?", storageType)
	}
	var rows []model.SysStorage
	err := q.Order("sort asc, id asc").Find(&rows).Error
	return rows, err
}

func (r *SysStorageRepository) GetByID(ctx context.Context, id any) (*model.SysStorage, error) {
	var row model.SysStorage
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysStorageRepository) GetByCode(ctx context.Context, code string) (*model.SysStorage, error) {
	var row model.SysStorage
	if err := r.DB.WithContext(ctx).Where("code = ?", code).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysStorageRepository) GetDefault(ctx context.Context) (*model.SysStorage, error) {
	var row model.SysStorage
	if err := r.DB.WithContext(ctx).Where("is_default = ? AND status = ?", true, "0").First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysStorageRepository) CountByIDs(ctx context.Context, ids []string) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&model.SysStorage{}).Where("id IN ?", ids).Count(&n).Error
	return n, err
}

func (r *SysStorageRepository) Create(ctx context.Context, row *model.SysStorage) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *SysStorageRepository) Save(ctx context.Context, row *model.SysStorage) error {
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *SysStorageRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	return r.DB.WithContext(ctx).Delete(&model.SysStorage{}, ids).Error
}

func (r *SysStorageRepository) ClearDefault(ctx context.Context) error {
	return r.DB.WithContext(ctx).Model(&model.SysStorage{}).Where("is_default = ?", true).Update("is_default", false).Error
}

func (r *SysStorageRepository) SetDefault(ctx context.Context, id int64) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SysStorage{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&model.SysStorage{}).Where("id = ?", id).Update("is_default", true).Error
	})
}

func (r *SysStorageRepository) ExistsCode(ctx context.Context, code string, excludeID int64) (bool, error) {
	q := r.DB.WithContext(ctx).Model(&model.SysStorage{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}
