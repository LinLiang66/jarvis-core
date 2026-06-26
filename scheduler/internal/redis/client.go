package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	redisconfig "jarvis-core/scheduler/internal/infra/redis"
)

const Prefix = "sched:"

type Client struct {
	raw *goredis.Client
}

func NewFromConfig(cfg redisconfig.Config) (*Client, error) {
	return dial(cfg.Addr, cfg.Password, cfg.DB, cfg.PoolSize, cfg.ReadTimeout)
}

func dial(addr, password string, db, poolSize int, readTimeout time.Duration) (*Client, error) {
	if poolSize <= 0 {
		poolSize = 10
	}
	if readTimeout <= 0 {
		readTimeout = 3 * time.Second
	}
	rdb := goredis.NewClient(&goredis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		ReadTimeout:  readTimeout,
		DialTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Client{raw: rdb}, nil
}

func (c *Client) Raw() *goredis.Client { return c.raw }

func (c *Client) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return c.raw.SetNX(ctx, Prefix+key, value, ttl).Result()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.raw.Get(ctx, Prefix+key).Result()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	prefixed := make([]string, len(keys))
	for i, k := range keys {
		prefixed[i] = Prefix + k
	}
	return c.raw.Del(ctx, prefixed...).Err()
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.raw.Exists(ctx, Prefix+key).Result()
	return n > 0, err
}

func (c *Client) Unlock(ctx context.Context, key, value string) error {
	const script = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`
	return c.raw.Eval(ctx, script, []string{Prefix + key}, value).Err()
}

func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.raw.RPush(ctx, Prefix+key, values...).Err()
}

func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return c.raw.LPop(ctx, Prefix+key).Result()
}

func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.raw.LLen(ctx, Prefix+key).Result()
}

func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.raw.Incr(ctx, Prefix+key).Result()
}

func InstanceScannerLockKey() string     { return "instance:scanner:lock" }
func RunningLockKey(jobID int64) string  { return fmt.Sprintf("running:lock:%d", jobID) }
func TriggerLockKey(jobID int64) string  { return fmt.Sprintf("trigger:lock:%d", jobID) }
func SerialQueueKey(jobID int64) string  { return fmt.Sprintf("serial:queue:%d", jobID) }
func RoundRobinKey(jobID int64) string   { return fmt.Sprintf("rr:%d", jobID) }
