package openplatform

import (
	"context"
	"sync"
	"time"

	"jarvis/backend/internal/pkg/logx"
	"jarvis/backend/internal/store"
)

const (
	syncMinute      = 5
	syncLockTTL     = 10 * time.Minute
	catchUpScanCount = 200
)

// StatSync 每小时 :05 将上一小时 Redis 统计同步到 MySQL（集群分布式锁 + 补偿扫描）。
type StatSync struct {
	store      *StatStore
	quotaStore *QuotaStore
	startOnce  sync.Once
}

func NewStatSync(store *StatStore, quotaStore *QuotaStore) *StatSync {
	return &StatSync{store: store, quotaStore: quotaStore}
}

func (s *StatSync) Enabled() bool {
	return s.store != nil && s.store.Enabled()
}

// Start 启动定时同步；同一进程仅启动一次。
func (s *StatSync) Start(ctx context.Context) {
	if !s.Enabled() {
		return
	}
	s.startOnce.Do(func() {
		go s.runLoop(ctx)
	})
}

func (s *StatSync) runLoop(ctx context.Context) {
	logx.Infof("[openplatform] stat sync started (hourly at :05, cluster lock enabled)")
	s.runCatchUp(ctx)

	timer := time.NewTimer(time.Until(nextSyncAt(time.Now())))
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-timer.C:
			s.syncPreviousHour(ctx, now)
			s.runCatchUp(ctx)
			timer.Reset(time.Until(nextSyncAt(now)))
		}
	}
}

// nextSyncAt 返回下一次整点过 5 分的触发时间。
func nextSyncAt(now time.Time) time.Time {
	loc := now.Location()
	next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), syncMinute, 0, 0, loc)
	if !now.Before(next) {
		next = next.Add(time.Hour)
	}
	return next
}

func previousHourKey(now time.Time) string {
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location()).
		Add(-time.Hour).Format("2006010215")
}

func (s *StatSync) syncPreviousHour(ctx context.Context, now time.Time) {
	hourKey := previousHourKey(now)
	if err := s.syncHour(ctx, hourKey); err != nil {
		logx.Infof("[openplatform] stat sync hour=%s failed: %v", hourKey, err)
	}
}

// runCatchUp 扫描 Redis 中已结束但未同步的小时桶，防止跨天或宕机漏同步。
func (s *StatSync) runCatchUp(ctx context.Context) {
	rdb := s.store.rdb
	if rdb == nil {
		return
	}
	keys, err := rdb.ScanKeys(ctx, statHourKeyPrefix+"*", catchUpScanCount)
	if err != nil {
		logx.Infof("[openplatform] stat catch-up scan failed: %v", err)
		return
	}
	cutoff := time.Now().Format("2006010215")
	for _, key := range keys {
		hourKey, ok := ParseHourKey(key)
		if !ok || hourKey >= cutoff {
			continue
		}
		if err := s.syncHour(ctx, hourKey); err != nil {
			logx.Infof("[openplatform] stat catch-up hour=%s failed: %v", hourKey, err)
		}
	}
}

func (s *StatSync) syncHour(ctx context.Context, hourKey string) error {
	if len(hourKey) != 10 {
		return nil
	}
	rdb := s.store.rdb
	stats := s.store.stats

	already, err := stats.IsHourSynced(ctx, hourKey)
	if err != nil {
		return err
	}
	if already {
		_ = rdb.Del(ctx, statHourKeyPrefix+hourKey)
		return nil
	}

	lockKey := syncLockKey(hourKey)
	token := newLockToken()
	acquired, err := rdb.SetNX(ctx, lockKey, token, syncLockTTL)
	if err != nil {
		return err
	}
	if !acquired {
		return nil
	}
	defer func() { _ = rdb.Unlock(ctx, lockKey, token) }()

	already, err = stats.IsHourSynced(ctx, hourKey)
	if err != nil {
		return err
	}
	if already {
		_ = rdb.Del(ctx, statHourKeyPrefix+hourKey)
		return nil
	}

	hKey := statHourKeyPrefix + hourKey
	fields, err := rdb.HGetAll(ctx, hKey)
	if err != nil {
		return err
	}
	entries := ParseHourHash(fields)
	statDate := HourKeyToStatDate(hourKey)
	if statDate == "" {
		return nil
	}

	synced, err := stats.SyncHourlyToDaily(ctx, hourKey, statDate, entries)
	if err != nil {
		return err
	}
	if synced || len(entries) == 0 {
		s.settleQuotaAfterSync(ctx, entries)
		_ = rdb.Del(ctx, hKey)
		if synced {
			logx.Infof("[openplatform] stat synced hour=%s date=%s entries=%d", hourKey, statDate, len(entries))
		}
	}
	return nil
}

// settleQuotaAfterSync 兜底将 Redis 余额同步到 MySQL（SET 余额，非再次扣减）。
func (s *StatSync) settleQuotaAfterSync(ctx context.Context, entries []store.HourlyStatEntry) {
	if s.quotaStore == nil || !s.quotaStore.Enabled() {
		return
	}
	appIDs := make(map[string]struct{})
	for _, e := range entries {
		appIDs[e.AppID] = struct{}{}
	}
	for appID := range appIDs {
		if err := s.quotaStore.FlushBalanceToDB(ctx, appID); err != nil {
			logx.Infof("[openplatform] quota sync flush balance app=%s err=%v", appID, err)
		}
	}
}
