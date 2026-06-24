package store

import (
	"context"
	"fmt"
	"time"

	infraredis "jarvis-core/backend/internal/infra/redis"
)

const authTokenKeyPrefix = "auth:token:"

// SessionCache Redis 会话缓存（登录 token）。
type SessionCache struct {
	rdb *infraredis.Client
	ttl time.Duration
}

func NewSessionCache(rdb *infraredis.Client, ttl time.Duration) *SessionCache {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &SessionCache{rdb: rdb, ttl: ttl}
}

func (s *SessionCache) Enabled() bool {
	return s != nil && s.rdb != nil && infraredis.Available()
}

func (s *SessionCache) SaveToken(ctx context.Context, token string, userID int64) error {
	if !s.Enabled() {
		return nil
	}
	return s.rdb.SetStr(ctx, authTokenKeyPrefix+token, fmt.Sprintf("%d", userID), s.ttl)
}

func (s *SessionCache) DeleteToken(ctx context.Context, token string) error {
	if !s.Enabled() {
		return nil
	}
	return s.rdb.Del(ctx, authTokenKeyPrefix+token)
}

func (s *SessionCache) ExistsToken(ctx context.Context, token string) (bool, error) {
	if !s.Enabled() {
		return true, nil
	}
	return s.rdb.Exists(ctx, authTokenKeyPrefix+token)
}
