package openplatform

import (
	"context"
	"fmt"

	"jarvis/backend/internal/model"
	"jarvis/backend/internal/pkg/logx"
	"jarvis/backend/internal/store"
)

// SyncActionRegistry 将代码注册的接口元数据同步到 MySQL 并生成文档。
func (s *Service) SyncActionRegistry(ctx context.Context) (int, error) {
	if s.actions == nil {
		return 0, nil
	}
	rows := RegistryToModels()
	if err := s.actions.UpsertFromRegistry(ctx, rows); err != nil {
		return 0, err
	}
	// 兜底：握手接口强制不计费（防止历史脏数据）
	_ = s.actions.UpdateBillableByActions(ctx, []string{ActionGetPublicKey, ActionCreateSecretKey}, false)
	logx.Infof("[openplatform] action registry synced count=%d", len(rows))
	return len(rows), nil
}

func (s *Service) ListActions(ctx context.Context, pq store.PageQuery, f store.OpenAPIActionFilter) ([]model.OpenAPIAction, int64, error) {
	return s.actions.List(ctx, pq, f)
}

func (s *Service) GetActionByID(ctx context.Context, id int64) (*model.OpenAPIAction, error) {
	return s.actions.GetByID(ctx, id)
}

func (s *Service) GetActionByAction(ctx context.Context, action string) (*model.OpenAPIAction, error) {
	return s.actions.GetByAction(ctx, action)
}

func (s *Service) CreateAction(ctx context.Context, row *model.OpenAPIAction) error {
	row.Source = "manual"
	RegenerateDoc(row)
	return s.actions.Create(ctx, row)
}

func (s *Service) UpdateAction(ctx context.Context, row *model.OpenAPIAction) error {
	existing, err := s.actions.GetByID(ctx, row.ID)
	if err != nil {
		return err
	}
	if existing.Source == "code" {
		// 代码注册接口：文档/schema/计费/加密 以代码注册表为准，仅允许改展示类字段
		title, category, desc, status, sort := row.Title, row.Category, row.Description, row.Status, row.Sort
		if meta, ok := GetRegisteredActionMeta(existing.Action); ok {
			applyRegistryMetaToRow(row, meta)
		} else {
			row.Encrypted = existing.Encrypted
			row.Billable = existing.Billable
			row.RequestSchema = existing.RequestSchema
			row.ResponseSchema = existing.ResponseSchema
			row.RequestFields = existing.RequestFields
			row.ResponseFields = existing.ResponseFields
		}
		row.ID = existing.ID
		row.Action = existing.Action
		row.Source = existing.Source
		row.Title = title
		row.Category = category
		row.Description = desc
		row.Status = status
		row.Sort = sort
		RegenerateDoc(row)
		return s.actions.Save(ctx, row)
	}
	RegenerateDoc(row)
	return s.actions.Save(ctx, row)
}

func (s *Service) DeleteActions(ctx context.Context, ids []string) error {
	return s.actions.DeleteByIDs(ctx, ids)
}

// PublicDocCategory 公开文档分类分组。
type PublicDocCategory struct {
	Name    string              `json:"name"`
	Actions []model.OpenAPIAction `json:"actions"`
}

// ListPublicDoc 返回启用状态的接口文档（按分类分组）。
func (s *Service) ListPublicDoc(ctx context.Context) ([]PublicDocCategory, error) {
	if s.actions == nil {
		return nil, nil
	}
	rows, err := s.actions.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	order := make([]string, 0)
	groups := map[string][]model.OpenAPIAction{}
	for _, row := range rows {
		cat := row.Category
		if cat == "" {
			cat = "其他"
		}
		if _, ok := groups[cat]; !ok {
			order = append(order, cat)
		}
		groups[cat] = append(groups[cat], row)
	}
	out := make([]PublicDocCategory, 0, len(order))
	for _, name := range order {
		out = append(out, PublicDocCategory{Name: name, Actions: groups[name]})
	}
	return out, nil
}

func (s *Service) GetPublicDocByAction(ctx context.Context, action string) (*model.OpenAPIAction, error) {
	row, err := s.actions.GetByAction(ctx, action)
	if err != nil {
		return nil, err
	}
	if row.Status != "0" {
		return nil, fmt.Errorf("action disabled")
	}
	return row, nil
}
