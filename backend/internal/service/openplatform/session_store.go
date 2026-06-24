package openplatform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	infraredis "jarvis/backend/internal/infra/redis"
	"jarvis/backend/internal/pkg/crypto"
)

const (
	sessionKeyPrefix = "open:session:"
	sessionTTL       = 2 * time.Hour
	// refreshThreshold 剩余 TTL 低于此值时续期（滑动过期，仅写 Redis 一次）。
	refreshThreshold = 30 * time.Minute
	// ttlCheckInterval 单节点对同一 token 探测 TTL 的最小间隔，降低 Redis 读压力。
	ttlCheckInterval = 5 * time.Minute
)

// SessionInfo 开放平台会话（Redis 集群共享的权威数据）。
type SessionInfo struct {
	AppID     string `json:"app_id"`
	Token     string `json:"token"`
	TDESKey   string `json:"tdes_key,omitempty"`
	CreatedAt int64  `json:"created_at"`
}

type sessionCacheEntry struct {
	cipher    *crypto.TDESCipher
	expiresAt time.Time
}

// SessionStore 会话与 3DES 加解密器管理。
// - Redis：集群权威源；滑动过期通过 EXPIRE 生效，任意节点续期全集群可见
// - 本地缓存：TDESCipher 懒加载；expiresAt 随 Redis 续期同步延长
type SessionStore struct {
	rdb *infraredis.Client

	mu           sync.RWMutex
	memSession   map[string]*SessionInfo
	memCryptors  map[string]*sessionCacheEntry
	lastTTLCheck sync.Map // token -> time.Time，仅用于节流 TTL 查询，非续期状态
}

func NewSessionStore(rdb *infraredis.Client) *SessionStore {
	return &SessionStore{
		rdb:         rdb,
		memSession:  make(map[string]*SessionInfo),
		memCryptors: make(map[string]*sessionCacheEntry),
	}
}

func (s *SessionStore) redisOK() bool {
	return s.rdb != nil && infraredis.Available()
}

func sessionRedisKey(token string) string {
	return sessionKeyPrefix + token
}

func (s *SessionStore) SaveToken(ctx context.Context, info *SessionInfo) error {
	if info.Token == "" {
		return fmt.Errorf("token required")
	}
	if s.redisOK() {
		return s.rdb.SetJSON(ctx, sessionRedisKey(info.Token), info, sessionTTL)
	}
	s.mu.Lock()
	s.memSession[info.Token] = cloneSession(info)
	s.mu.Unlock()
	return nil
}

func (s *SessionStore) GetByToken(ctx context.Context, token string) (*SessionInfo, error) {
	return s.loadSession(ctx, token)
}

// SaveTDESKey 持久化 3DES 密钥到 Redis，并在当前节点预热 TDESCipher（仅初始化一次）。
func (s *SessionStore) SaveTDESKey(ctx context.Context, token, key string) error {
	info, err := s.loadSession(ctx, token)
	if err != nil {
		return err
	}
	info.TDESKey = key
	if err := s.SaveToken(ctx, info); err != nil {
		return err
	}
	_, err = s.buildAndCacheCryptor(token, key)
	return err
}

// TouchSession 滑动续期：仅在剩余 TTL 不足时写 Redis，集群内任意节点执行即可全局生效。
func (s *SessionStore) TouchSession(ctx context.Context, token string) {
	if token == "" {
		return
	}
	if !s.redisOK() {
		s.touchSessionMemory(token)
		return
	}
	if !s.shouldCheckTTL(token) {
		return
	}
	remaining, err := s.rdb.TTL(ctx, sessionRedisKey(token))
	if err != nil || remaining <= 0 {
		return
	}
	if remaining < refreshThreshold {
		_ = s.rdb.SetExpire(ctx, sessionRedisKey(token), sessionTTL)
		s.extendLocalCryptorExpiry(token)
	}
}

func (s *SessionStore) shouldCheckTTL(token string) bool {
	now := time.Now()
	if v, ok := s.lastTTLCheck.Load(token); ok {
		if now.Sub(v.(time.Time)) < ttlCheckInterval {
			return false
		}
	}
	s.lastTTLCheck.Store(token, now)
	return true
}

func (s *SessionStore) touchSessionMemory(token string) {
	if !s.shouldCheckTTL(token) {
		return
	}
	s.mu.RLock()
	entry, ok := s.memCryptors[token]
	s.mu.RUnlock()
	if !ok {
		return
	}
	if time.Until(entry.expiresAt) < refreshThreshold {
		s.extendLocalCryptorExpiry(token)
	}
}

func (s *SessionStore) extendLocalCryptorExpiry(token string) {
	s.mu.Lock()
	if entry, ok := s.memCryptors[token]; ok {
		entry.expiresAt = time.Now().Add(sessionTTL)
	}
	s.mu.Unlock()
}

// GetCryptor 获取可复用加解密器：本地命中直接返回；未命中从 Redis 加载密钥并初始化一次。
func (s *SessionStore) GetCryptor(ctx context.Context, token string) (*crypto.TDESCipher, error) {
	if c := s.getLocalCryptor(token); c != nil {
		return c, nil
	}
	info, err := s.loadSession(ctx, token)
	if err != nil {
		return nil, err
	}
	if info.TDESKey == "" {
		return nil, fmt.Errorf("3des key not initialized")
	}
	return s.buildAndCacheCryptor(token, info.TDESKey)
}

func (s *SessionStore) loadSession(ctx context.Context, token string) (*SessionInfo, error) {
	if token == "" {
		return nil, fmt.Errorf("token required")
	}
	if s.redisOK() {
		var info SessionInfo
		if err := s.rdb.GetJSON(ctx, sessionRedisKey(token), &info); err != nil {
			return nil, fmt.Errorf("token not found")
		}
		return &info, nil
	}
	s.mu.RLock()
	info, ok := s.memSession[token]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("token not found")
	}
	return cloneSession(info), nil
}

func (s *SessionStore) buildAndCacheCryptor(token, key string) (*crypto.TDESCipher, error) {
	cipher, err := crypto.NewTDESCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	entry := &sessionCacheEntry{
		cipher:    cipher,
		expiresAt: time.Now().Add(sessionTTL),
	}
	s.mu.Lock()
	s.memCryptors[token] = entry
	s.mu.Unlock()
	return cipher, nil
}

func (s *SessionStore) getLocalCryptor(token string) *crypto.TDESCipher {
	s.mu.RLock()
	entry, ok := s.memCryptors[token]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	if time.Now().After(entry.expiresAt) {
		s.mu.Lock()
		delete(s.memCryptors, token)
		s.mu.Unlock()
		return nil
	}
	return entry.cipher
}

func cloneSession(info *SessionInfo) *SessionInfo {
	if info == nil {
		return nil
	}
	cp := *info
	return &cp
}

func (s *SessionStore) MarshalSession(info *SessionInfo) string {
	b, _ := json.Marshal(info)
	return string(b)
}

// Enabled 是否启用 Redis 集群会话（生产环境建议开启）。
func (s *SessionStore) Enabled() bool {
	return s.redisOK()
}
