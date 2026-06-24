package store

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"jarvis/backend/internal/infra/base"
	"jarvis/backend/internal/model"
)

type OpenAppFilter struct {
	AppID   string
	AppName string
	Status  string
}

type OpenAppRepository struct {
	base.CRUD
}

func NewOpenAppRepository(db *gorm.DB) *OpenAppRepository {
	return &OpenAppRepository{CRUD: base.CRUD{DB: db}}
}

func (r *OpenAppRepository) AutoMigrate(ctx context.Context) error {
	return r.DB.WithContext(ctx).AutoMigrate(&model.OpenApp{})
}

func (r *OpenAppRepository) List(ctx context.Context, pq PageQuery, f OpenAppFilter) ([]model.OpenApp, int64, error) {
	return ListPage[model.OpenApp](ctx, r.DB, pq, func(q *gorm.DB) *gorm.DB {
		if f.AppID != "" {
			q = q.Where("app_id LIKE ?", "%"+f.AppID+"%")
		}
		if f.AppName != "" {
			q = q.Where("app_name LIKE ?", "%"+f.AppName+"%")
		}
		if f.Status != "" {
			q = q.Where("status = ?", f.Status)
		}
		return q
	})
}

func (r *OpenAppRepository) GetByID(ctx context.Context, id any) (*model.OpenApp, error) {
	var row model.OpenApp
	if err := r.DB.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OpenAppRepository) GetByAppID(ctx context.Context, appID string) (*model.OpenApp, error) {
	var row model.OpenApp
	if err := r.DB.WithContext(ctx).Where("app_id = ? AND status = ?", appID, "0").First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OpenAppRepository) Create(ctx context.Context, row *model.OpenApp) error {
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *OpenAppRepository) Save(ctx context.Context, row *model.OpenApp) error {
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *OpenAppRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	return r.DB.WithContext(ctx).Delete(&model.OpenApp{}, ids).Error
}

func (r *OpenAppRepository) TryDeductQuota(ctx context.Context, appID string) (bool, error) {
	res := r.DB.WithContext(ctx).Model(&model.OpenApp{}).
		Where("app_id = ? AND total_quota > 0", appID).
		UpdateColumn("total_quota", gorm.Expr("total_quota - 1"))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *OpenAppRepository) RefundQuota(ctx context.Context, appID string) error {
	return r.DB.WithContext(ctx).Model(&model.OpenApp{}).
		Where("app_id = ?", appID).
		UpdateColumn("total_quota", gorm.Expr("total_quota + 1")).Error
}

func (r *OpenAppRepository) ConfirmCall(ctx context.Context, appID string) error {
	return r.DB.WithContext(ctx).Model(&model.OpenApp{}).
		Where("app_id = ?", appID).
		UpdateColumn("total_calls", gorm.Expr("total_calls + 1")).Error
}

func (r *OpenAppRepository) SetQuotaBalance(ctx context.Context, appID string, balance int) error {
	return r.DB.WithContext(ctx).Model(&model.OpenApp{}).
		Where("app_id = ?", appID).
		UpdateColumn("total_quota", balance).Error
}

func (r *OpenAppRepository) IncrTotalCalls(ctx context.Context, appID string, delta int64) error {
	if delta <= 0 {
		return nil
	}
	return r.DB.WithContext(ctx).Model(&model.OpenApp{}).
		Where("app_id = ?", appID).
		UpdateColumn("total_calls", gorm.Expr("total_calls + ?", delta)).Error
}

type OpenAPIStatFilter struct {
	AppID    string
	Action   string
	StatDate string
	DateFrom string
	DateTo   string
}

type OpenAPIStatRepository struct {
	base.CRUD
}

func NewOpenAPIStatRepository(db *gorm.DB) *OpenAPIStatRepository {
	return &OpenAPIStatRepository{CRUD: base.CRUD{DB: db}}
}

func (r *OpenAPIStatRepository) AutoMigrate(ctx context.Context) error {
	return errorsJoinMigrate(
		r.DB.WithContext(ctx).AutoMigrate(&model.OpenAPICallLog{}),
		r.DB.WithContext(ctx).AutoMigrate(&model.OpenAPIDailyStat{}),
		r.DB.WithContext(ctx).AutoMigrate(&model.OpenAPIHourlySyncLog{}),
	)
}

func errorsJoinMigrate(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func (r *OpenAPIStatRepository) ListDailyStat(ctx context.Context, pq PageQuery, f OpenAPIStatFilter) ([]model.OpenAPIDailyStat, int64, error) {
	return ListPage[model.OpenAPIDailyStat](ctx, r.DB, pq, func(q *gorm.DB) *gorm.DB {
		if f.AppID != "" {
			q = q.Where("app_id LIKE ?", "%"+f.AppID+"%")
		}
		if f.Action != "" {
			q = q.Where("action LIKE ?", "%"+f.Action+"%")
		}
		if f.StatDate != "" {
			q = q.Where("stat_date = ?", f.StatDate)
		}
		if f.DateFrom != "" {
			q = q.Where("stat_date >= ?", f.DateFrom)
		}
		if f.DateTo != "" {
			q = q.Where("stat_date <= ?", f.DateTo)
		}
		return q.Order("stat_date DESC")
	})
}

func (r *OpenAPIStatRepository) ListCallLogs(ctx context.Context, pq PageQuery, f OpenAPIStatFilter) ([]model.OpenAPICallLog, int64, error) {
	return ListPage[model.OpenAPICallLog](ctx, r.DB, pq, func(q *gorm.DB) *gorm.DB {
		if f.AppID != "" {
			q = q.Where("app_id LIKE ?", "%"+f.AppID+"%")
		}
		if f.Action != "" {
			q = q.Where("action LIKE ?", "%"+f.Action+"%")
		}
		if f.DateFrom != "" {
			q = q.Where("created_at >= ?", f.DateFrom)
		}
		if f.DateTo != "" {
			to := f.DateTo
			if len(to) == 10 {
				to += " 23:59:59"
			}
			q = q.Where("created_at <= ?", to)
		}
		return q.Order("created_at DESC")
	})
}

func (r *OpenAPIStatRepository) RecordCallLog(ctx context.Context, log *model.OpenAPICallLog) error {
	return r.DB.WithContext(ctx).Create(log).Error
}

// RecordCallStatDirect 无 Redis 时的降级：直接写入日统计。
func (r *OpenAPIStatRepository) RecordCallStatDirect(ctx context.Context, appID, action string, success bool) error {
	statDate := time.Now().Format("2006-01-02")
	successDelta := int64(0)
	failDelta := int64(0)
	if success {
		successDelta = 1
	} else {
		failDelta = 1
	}
	return r.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "app_id"}, {Name: "action"}, {Name: "stat_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_count":   gorm.Expr("total_count + 1"),
			"success_count": gorm.Expr("success_count + ?", successDelta),
			"fail_count":    gorm.Expr("fail_count + ?", failDelta),
		}),
	}).Create(&model.OpenAPIDailyStat{
		AppID:        appID,
		Action:       action,
		StatDate:     statDate,
		TotalCount:   1,
		SuccessCount: successDelta,
		FailCount:    failDelta,
	}).Error
}

// HourlyStatEntry 单小时桶内的一条聚合统计。
type HourlyStatEntry struct {
	AppID        string
	Action       string
	TotalCount   int64
	SuccessCount int64
	FailCount    int64
}

// SyncHourlyToDaily 将某小时桶数据幂等合并到日统计表；返回 false 表示该小时已同步过。
func (r *OpenAPIStatRepository) SyncHourlyToDaily(ctx context.Context, hourKey, statDate string, entries []HourlyStatEntry) (bool, error) {
	if len(entries) == 0 {
		return r.markHourSynced(ctx, hourKey)
	}
	var synced bool
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		logRow := &model.OpenAPIHourlySyncLog{HourKey: hourKey}
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(logRow)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		for _, e := range entries {
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "app_id"}, {Name: "action"}, {Name: "stat_date"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"total_count":   gorm.Expr("total_count + ?", e.TotalCount),
					"success_count": gorm.Expr("success_count + ?", e.SuccessCount),
					"fail_count":    gorm.Expr("fail_count + ?", e.FailCount),
				}),
			}).Create(&model.OpenAPIDailyStat{
				AppID:        e.AppID,
				Action:       e.Action,
				StatDate:     statDate,
				TotalCount:   e.TotalCount,
				SuccessCount: e.SuccessCount,
				FailCount:    e.FailCount,
			}).Error; err != nil {
				return err
			}
		}
		synced = true
		return nil
	})
	return synced, err
}

func (r *OpenAPIStatRepository) markHourSynced(ctx context.Context, hourKey string) (bool, error) {
	logRow := &model.OpenAPIHourlySyncLog{HourKey: hourKey}
	res := r.DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(logRow)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *OpenAPIStatRepository) IsHourSynced(ctx context.Context, hourKey string) (bool, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&model.OpenAPIHourlySyncLog{}).
		Where("hour_key = ?", hourKey).Count(&n).Error
	return n > 0, err
}
