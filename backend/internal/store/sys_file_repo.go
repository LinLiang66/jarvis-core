package store

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"jarvis-core/backend/internal/infra/base"
	"jarvis-core/backend/internal/model"
)

type SysFileFilter struct {
	StorageID    int64
	ParentPath   string
	OriginalName string
	Type         *int
}

type SysFileRepository struct{ base.CRUD }

func NewSysFileRepository(db *gorm.DB) *SysFileRepository {
	return &SysFileRepository{CRUD: base.CRUD{DB: db}}
}

func (r *SysFileRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.SysFile{})
}

func (r *SysFileRepository) List(ctx context.Context, pq PageQuery, f SysFileFilter) ([]model.SysFile, int64, error) {
	return ListPage[model.SysFile](ctx, r.DB, pq, func(q *gorm.DB) *gorm.DB {
		if f.StorageID > 0 {
			q = q.Where("storage_id = ?", f.StorageID)
		}
		if f.ParentPath != "" {
			q = q.Where("parent_path = ?", f.ParentPath)
		}
		if f.OriginalName != "" {
			q = q.Where("original_name LIKE ?", "%"+f.OriginalName+"%")
		}
		if f.Type != nil {
			q = q.Where("type = ?", *f.Type)
		}
		return q.Order("type asc, id desc")
	})
}

func (r *SysFileRepository) GetByID(ctx context.Context, id any) (*model.SysFile, error) {
	var row model.SysFile
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SysFileRepository) CountByStorageIDs(ctx context.Context, ids []string) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&model.SysFile{}).Where("storage_id IN ?", ids).Count(&n).Error
	return n, err
}

func (r *SysFileRepository) Create(ctx context.Context, row *model.SysFile) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *SysFileRepository) Save(ctx context.Context, row *model.SysFile) error {
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *SysFileRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	return r.DB.WithContext(ctx).Delete(&model.SysFile{}, ids).Error
}

func (r *SysFileRepository) Statistics(ctx context.Context) (fileCount int64, dirCount int64, totalSize int64, err error) {
	err = r.DB.WithContext(ctx).Model(&model.SysFile{}).Where("type = ?", 1).Count(&fileCount).Error
	if err != nil {
		return
	}
	err = r.DB.WithContext(ctx).Model(&model.SysFile{}).Where("type = ?", 0).Count(&dirCount).Error
	if err != nil {
		return
	}
	err = r.DB.WithContext(ctx).Model(&model.SysFile{}).Where("type = ?", 1).Select("COALESCE(SUM(size),0)").Scan(&totalSize).Error
	return
}

func (r *SysFileRepository) ExistsDir(ctx context.Context, storageID int64, parentPath, name string) (bool, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&model.SysFile{}).
		Where("storage_id = ? AND parent_path = ? AND original_name = ? AND type = ?", storageID, parentPath, name, 0).
		Count(&n).Error
	return n > 0, err
}

// ListByPathPrefix 查询目录及其全部子孙项（含自身）。
func (r *SysFileRepository) ListByPathPrefix(ctx context.Context, storageID int64, pathPrefix string) ([]model.SysFile, error) {
	pathPrefix = strings.TrimSpace(pathPrefix)
	var rows []model.SysFile
	err := r.DB.WithContext(ctx).
		Where("storage_id = ? AND (path = ? OR path LIKE ?)", storageID, pathPrefix, pathPrefix+"/%").
		Order("type desc, length(path) desc, id desc").
		Find(&rows).Error
	return rows, err
}
