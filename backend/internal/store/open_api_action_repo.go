package store

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jarvis/backend/internal/infra/base"
	"jarvis/backend/internal/model"
)

type OpenAPIActionFilter struct {
	Action   string
	Title    string
	Category string
	Status   string
}

type OpenAPIActionRepository struct {
	base.CRUD
}

func NewOpenAPIActionRepository(db *gorm.DB) *OpenAPIActionRepository {
	return &OpenAPIActionRepository{CRUD: base.CRUD{DB: db}}
}

func (r *OpenAPIActionRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.OpenAPIAction{})
}

func (r *OpenAPIActionRepository) List(ctx context.Context, pq PageQuery, f OpenAPIActionFilter) ([]model.OpenAPIAction, int64, error) {
	return ListPage[model.OpenAPIAction](ctx, r.DB, pq, func(q *gorm.DB) *gorm.DB {
		if f.Action != "" {
			q = q.Where("action LIKE ?", "%"+f.Action+"%")
		}
		if f.Title != "" {
			q = q.Where("title LIKE ?", "%"+f.Title+"%")
		}
		if f.Category != "" {
			q = q.Where("category = ?", f.Category)
		}
		if f.Status != "" {
			q = q.Where("status = ?", f.Status)
		}
		return q.Order("sort ASC, id ASC")
	})
}

func (r *OpenAPIActionRepository) GetByID(ctx context.Context, id any) (*model.OpenAPIAction, error) {
	var row model.OpenAPIAction
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OpenAPIActionRepository) GetByAction(ctx context.Context, action string) (*model.OpenAPIAction, error) {
	var row model.OpenAPIAction
	if err := r.DB.WithContext(ctx).Where("action = ?", action).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OpenAPIActionRepository) Create(ctx context.Context, row *model.OpenAPIAction) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *OpenAPIActionRepository) Save(ctx context.Context, row *model.OpenAPIAction) error {
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *OpenAPIActionRepository) ListEnabled(ctx context.Context) ([]model.OpenAPIAction, error) {
	var rows []model.OpenAPIAction
	err := r.DB.WithContext(ctx).Where("status = ?", "0").
		Order("sort ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *OpenAPIActionRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	return r.DB.WithContext(ctx).Delete(&model.OpenAPIAction{}, ids).Error
}

func (r *OpenAPIActionRepository) UpdateBillableByActions(ctx context.Context, actions []string, billable bool) error {
	if len(actions) == 0 {
		return nil
	}
	return r.DB.WithContext(ctx).Model(&model.OpenAPIAction{}).
		Where("action IN ?", actions).
		Update("billable", billable).Error
}

// UpsertFromRegistry 代码注册表同步入库（按 action 幂等 upsert）。
func (r *OpenAPIActionRepository) UpsertFromRegistry(ctx context.Context, rows []model.OpenAPIAction) error {
	for i := range rows {
		row := &rows[i]
		// 必须用 map 赋值：struct 中 billable/encrypted=false 为零值，GORM upsert 会跳过导致无法更新为「否」
		err := r.DB.WithContext(ctx).
			Select("*").
			Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "action"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":           row.Title,
				"category":        row.Category,
				"description":     row.Description,
				"encrypted":       row.Encrypted,
				"billable":        row.Billable,
				"request_schema":  row.RequestSchema,
				"response_schema": row.ResponseSchema,
				"request_fields":  row.RequestFields,
				"response_fields": row.ResponseFields,
				"doc_markdown":    row.DocMarkdown,
				"sort":            row.Sort,
				"source":          row.Source,
			}),
		}).Create(row).Error
		if err != nil {
			return err
		}
	}
	return nil
}
