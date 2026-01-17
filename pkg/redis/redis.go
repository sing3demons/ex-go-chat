package redis

import (
	"context"
	"fmt"
	"time"

	"realtime-chat-system/config"

	"github.com/redis/go-redis/v9"
)

// Client wraps Redis client with additional functionality
type Client struct {
	rdb *redis.Client
	ctx context.Context
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	ctx := context.Background()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	fmt.Println("Redis client connected successfully")

	return &Client{
		rdb: rdb,
		ctx: ctx,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Set sets a key-value pair with expiration
func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(c.ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (c *Client) Get(key string) (string, error) {
	return c.rdb.Get(c.ctx, key).Result()
}

// Del deletes keys
func (c *Client) Del(keys ...string) error {
	return c.rdb.Del(c.ctx, keys...).Err()
}

// Exists checks if key exists
func (c *Client) Exists(key string) (bool, error) {
	result, err := c.rdb.Exists(c.ctx, key).Result()
	return result > 0, err
}

// Expire sets expiration for a key
func (c *Client) Expire(key string, expiration time.Duration) error {
	return c.rdb.Expire(c.ctx, key, expiration).Err()
}

// HSet sets hash field
func (c *Client) HSet(key string, field string, value interface{}) error {
	return c.rdb.HSet(c.ctx, key, field, value).Err()
}

// HGet gets hash field
func (c *Client) HGet(key string, field string) (string, error) {
	return c.rdb.HGet(c.ctx, key, field).Result()
}

// HGetAll gets all hash fields
func (c *Client) HGetAll(key string) (map[string]string, error) {
	return c.rdb.HGetAll(c.ctx, key).Result()
}

// HDel deletes hash fields
func (c *Client) HDel(key string, fields ...string) error {
	return c.rdb.HDel(c.ctx, key, fields...).Err()
}

// SAdd adds members to set
func (c *Client) SAdd(key string, members ...interface{}) error {
	return c.rdb.SAdd(c.ctx, key, members...).Err()
}

// SRem removes members from set
func (c *Client) SRem(key string, members ...interface{}) error {
	return c.rdb.SRem(c.ctx, key, members...).Err()
}

// SMembers gets all set members
func (c *Client) SMembers(key string) ([]string, error) {
	return c.rdb.SMembers(c.ctx, key).Result()
}

// SIsMember checks if member exists in set
func (c *Client) SIsMember(key string, member interface{}) (bool, error) {
	return c.rdb.SIsMember(c.ctx, key, member).Result()
}

// Publish publishes message to channel
func (c *Client) Publish(channel string, message interface{}) error {
	return c.rdb.Publish(c.ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (c *Client) Subscribe(channels ...string) *redis.PubSub {
	return c.rdb.Subscribe(c.ctx, channels...)
}

// ZAdd adds members to sorted set
func (c *Client) ZAdd(key string, members ...redis.Z) error {
	return c.rdb.ZAdd(c.ctx, key, members...).Err()
}

// ZRem removes members from sorted set
func (c *Client) ZRem(key string, members ...interface{}) error {
	return c.rdb.ZRem(c.ctx, key, members...).Err()
}

// ZRange gets sorted set members by range
func (c *Client) ZRange(key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRange(c.ctx, key, start, stop).Result()
}

// ZRevRange gets sorted set members by reverse range
func (c *Client) ZRevRange(key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRevRange(c.ctx, key, start, stop).Result()
}

// GetClient returns the underlying Redis client for advanced operations
func (c *Client) GetClient() *redis.Client {
	return c.rdb
}