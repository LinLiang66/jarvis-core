package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client 参考 tiny-pro-go/utils.RedisUtil 的封装。
type Client struct {
	raw *goredis.Client
}

var (
	global      *Client
	once        sync.Once
	initErr     error
	initialized bool
)

func Init(cfg Config) (*Client, error) {
	return InitWithContext(context.Background(), cfg)
}

func InitWithContext(ctx context.Context, cfg Config) (*Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	once.Do(func() {
		rdb := goredis.NewClient(&goredis.Options{
			Addr:         cfg.Addr,
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			ReadTimeout:  cfg.ReadTimeout,
			DialTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		})
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := rdb.Ping(pingCtx).Err(); err != nil {
			initErr = fmt.Errorf("redis ping %s: %w", cfg.Addr, err)
			_ = rdb.Close()
			return
		}
		global = &Client{raw: rdb}
		initialized = true
	})
	if initErr != nil {
		return nil, initErr
	}
	return global, nil
}

func Must() *Client {
	if !initialized {
		panic("redis not initialized")
	}
	return global
}

func Available() bool {
	return initialized && global != nil
}

func (c *Client) Raw() *goredis.Client {
	return c.raw
}

func (c *Client) SetStr(ctx context.Context, key, value string, expiration time.Duration) error {
	return c.raw.Set(ctx, key, value, expiration).Err()
}

func (c *Client) GetStr(ctx context.Context, key string) (string, error) {
	return c.raw.Get(ctx, key).Result()
}

func (c *Client) SetJSON(ctx context.Context, key string, v any, expiration time.Duration) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.SetStr(ctx, key, string(raw), expiration)
}

func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
	val, err := c.GetStr(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.raw.Del(ctx, keys...).Err()
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.raw.Exists(ctx, key).Result()
	return n > 0, err
}

func (c *Client) SetExpire(ctx context.Context, key string, expiration time.Duration) error {
	return c.raw.Expire(ctx, key, expiration).Err()
}

// TTL 返回 key 剩余过期时间；-2 表示不存在，-1 表示无过期时间。
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.raw.TTL(ctx, key).Result()
}

func (c *Client) HIncrBy(ctx context.Context, key, field string, delta int64) (int64, error) {
	return c.raw.HIncrBy(ctx, key, field, delta).Result()
}

func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.raw.HGetAll(ctx, key).Result()
}

// SetNX 仅在 key 不存在时写入，常用于分布式锁。
func (c *Client) SetNX(ctx context.Context, key, value string, expiration time.Duration) (bool, error) {
	return c.raw.SetNX(ctx, key, value, expiration).Result()
}

// Unlock 仅当 value 匹配时删除 key（安全释放锁）。
func (c *Client) Unlock(ctx context.Context, key, value string) error {
	const script = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
	return c.raw.Eval(ctx, script, []string{key}, value).Err()
}

// ScanKeys 按 pattern 扫描 key（非阻塞，适合补偿任务）。
func (c *Client) ScanKeys(ctx context.Context, pattern string, count int64) ([]string, error) {
	var keys []string
	iter := c.raw.Scan(ctx, 0, pattern, count).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	return keys, iter.Err()
}

func (c *Client) IncrBy(ctx context.Context, key string, delta int64) (int64, error) {
	return c.raw.IncrBy(ctx, key, delta).Result()
}

func (c *Client) EvalInt64(ctx context.Context, script string, keys []string, args ...interface{}) (int64, error) {
	return c.raw.Eval(ctx, script, keys, args...).Int64()
}
