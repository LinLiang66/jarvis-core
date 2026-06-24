package openplatform

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	infraredis "jarvis/backend/internal/infra/redis"
	"jarvis/backend/internal/store"
)

const (
	statHourKeyPrefix = "open:stat:hour:"
	statSyncLockPref  = "open:stat:sync:lock:"
	statHourTTL       = 72 * time.Hour
	statFieldSep      = "|"
)

// StatStore 调用统计 Redis 实时计数；无 Redis 时降级写 MySQL。
type StatStore struct {
	rdb   *infraredis.Client
	stats *store.OpenAPIStatRepository
}

func NewStatStore(rdb *infraredis.Client, stats *store.OpenAPIStatRepository) *StatStore {
	return &StatStore{rdb: rdb, stats: stats}
}

func (s *StatStore) Enabled() bool {
	return s.rdb != nil && infraredis.Available()
}

func hourBucketKey(t time.Time) string {
	return statHourKeyPrefix + t.Format("2006010215")
}

func statField(appID, action, metric string) string {
	return appID + statFieldSep + action + statFieldSep + metric
}

// Record 记录一次 API 调用到 Redis 小时桶。
func (s *StatStore) Record(ctx context.Context, appID, action string, success bool) error {
	if appID == "" || action == "" {
		return nil
	}
	if !s.Enabled() {
		return s.stats.RecordCallStatDirect(ctx, appID, action, success)
	}
	now := time.Now()
	hKey := hourBucketKey(now)
	totalField := statField(appID, action, "t")
	successField := statField(appID, action, "s")
	failField := statField(appID, action, "f")

	if _, err := s.rdb.HIncrBy(ctx, hKey, totalField, 1); err != nil {
		return err
	}
	if success {
		if _, err := s.rdb.HIncrBy(ctx, hKey, successField, 1); err != nil {
			return err
		}
	} else if _, err := s.rdb.HIncrBy(ctx, hKey, failField, 1); err != nil {
		return err
	}
	_ = s.rdb.SetExpire(ctx, hKey, statHourTTL)
	return nil
}

// ParseHourKey 从 Redis key 解析 YYYYMMDDHH。
func ParseHourKey(redisKey string) (string, bool) {
	if !strings.HasPrefix(redisKey, statHourKeyPrefix) {
		return "", false
	}
	hourKey := strings.TrimPrefix(redisKey, statHourKeyPrefix)
	if len(hourKey) != 10 {
		return "", false
	}
	return hourKey, true
}

// HourKeyToStatDate 将 YYYYMMDDHH 转为日统计日期 YYYY-MM-DD。
func HourKeyToStatDate(hourKey string) string {
	if len(hourKey) < 8 {
		return ""
	}
	return hourKey[:4] + "-" + hourKey[4:6] + "-" + hourKey[6:8]
}

// ParseHourHash 解析小时桶 HGETALL 结果为入库条目。
func ParseHourHash(fields map[string]string) []store.HourlyStatEntry {
	type agg struct {
		total, success, fail int64
	}
	merged := make(map[string]*agg)
	for field, val := range fields {
		parts := strings.Split(field, statFieldSep)
		if len(parts) != 3 {
			continue
		}
		appID, action, metric := parts[0], parts[1], parts[2]
		key := appID + statFieldSep + action
		if merged[key] == nil {
			merged[key] = &agg{}
		}
		var n int64
		if _, err := fmt.Sscan(val, &n); err != nil {
			continue
		}
		switch metric {
		case "t":
			merged[key].total = n
		case "s":
			merged[key].success = n
		case "f":
			merged[key].fail = n
		}
	}
	out := make([]store.HourlyStatEntry, 0, len(merged))
	for key, a := range merged {
		parts := strings.SplitN(key, statFieldSep, 2)
		if len(parts) != 2 || a.total == 0 {
			continue
		}
		out = append(out, store.HourlyStatEntry{
			AppID:        parts[0],
			Action:       parts[1],
			TotalCount:   a.total,
			SuccessCount: a.success,
			FailCount:    a.fail,
		})
	}
	return out
}

func syncLockKey(hourKey string) string {
	return statSyncLockPref + hourKey
}

func newLockToken() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
