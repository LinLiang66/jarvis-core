package store

import (
	"context"

	"gorm.io/gorm"
)

type PageQuery struct {
	Page  int
	Size  int
	Order string
}

func NormalizePage(page, size int) (offset, limit int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	return (page - 1) * size, size
}

// ListPage 通用分页查询。
func ListPage[T any](ctx context.Context, db *gorm.DB, pq PageQuery, scopes ...func(*gorm.DB) *gorm.DB) ([]T, int64, error) {
	q := db.WithContext(ctx).Model(new(T))
	for _, scope := range scopes {
		if scope != nil {
			q = scope(q)
		}
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset, limit := NormalizePage(pq.Page, pq.Size)
	order := pq.Order
	if order == "" {
		order = "id desc"
	}
	var list []T
	if err := q.Order(order).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
