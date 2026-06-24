package openplatform

import (
	"context"
	"strconv"

	infraredis "jarvis/backend/internal/infra/redis"
	"jarvis/backend/internal/store"
)

const quotaBalPrefix = "open:quota:bal:"

// QuotaStore Redis 实时配额余额；MySQL 由定时同步落库，避免请求路径与统计同步双重扣减。
type QuotaStore struct {
	rdb  *infraredis.Client
	apps *store.OpenAppRepository
}

func NewQuotaStore(rdb *infraredis.Client, apps *store.OpenAppRepository) *QuotaStore {
	return &QuotaStore{rdb: rdb, apps: apps}
}

func (q *QuotaStore) Enabled() bool {
	return q.rdb != nil && infraredis.Available()
}

func (q *QuotaStore) balanceKey(appID string) string {
	return quotaBalPrefix + appID
}

// SyncBalanceFromDB 将 MySQL 余额写入 Redis（创建/充值/编辑配额时调用）。
func (q *QuotaStore) SyncBalanceFromDB(ctx context.Context, appID string, balance int) error {
	if !q.Enabled() {
		return nil
	}
	return q.rdb.SetStr(ctx, q.balanceKey(appID), strconv.Itoa(balance), 0)
}

func (q *QuotaStore) ensureInit(ctx context.Context, appID string) error {
	if !q.Enabled() {
		return nil
	}
	ok, err := q.rdb.Exists(ctx, q.balanceKey(appID))
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	app, err := q.apps.GetByAppID(ctx, appID)
	if err != nil {
		return err
	}
	return q.SyncBalanceFromDB(ctx, appID, app.TotalQuota)
}

// TryDeduct 原子预扣 Redis 余额。
func (q *QuotaStore) TryDeduct(ctx context.Context, appID string) (bool, error) {
	if err := q.ensureInit(ctx, appID); err != nil {
		return false, err
	}
	const script = `
local v = tonumber(redis.call('GET', KEYS[1]) or '-1')
if v <= 0 then return 0 end
redis.call('DECR', KEYS[1])
return 1`
	n, err := q.rdb.EvalInt64(ctx, script, []string{q.balanceKey(appID)})
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// Refund 失败退还 Redis 余额。
func (q *QuotaStore) Refund(ctx context.Context, appID string) error {
	if !q.Enabled() {
		return nil
	}
	_, err := q.rdb.IncrBy(ctx, q.balanceKey(appID), 1)
	return err
}

// FlushBalanceToDB 将 Redis 余额同步到 MySQL（定时任务调用，非再次扣减）。
func (q *QuotaStore) FlushBalanceToDB(ctx context.Context, appID string) error {
	if !q.Enabled() {
		return nil
	}
	val, err := q.rdb.GetStr(ctx, q.balanceKey(appID))
	if err != nil {
		return err
	}
	balance, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	return q.apps.SetQuotaBalance(ctx, appID, balance)
}
