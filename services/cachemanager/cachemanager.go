package cachemanager

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// RedisClient defines an interface for mocking Redis operations
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

// CacheManager handles caching operations
type CacheManager struct {
	rdb RedisClient
	log zerolog.Logger
}

// NewCacheManager initializes a new CacheManager
func NewCacheManager(rdb RedisClient, log zerolog.Logger) *CacheManager {
	return &CacheManager{
		rdb: rdb,
		log: log,
	}
}

// Set stores a value in Redis with TTL in minutes
func (cm *CacheManager) Set(ctx context.Context, key string, value string, ttl int) error {
	return cm.rdb.Set(ctx, key, value, time.Duration(ttl)*time.Minute).Err()
}

// Get retrieves a value from Redis
func (cm *CacheManager) Get(ctx context.Context, key string) (string, error) {
	return cm.rdb.Get(ctx, key).Result()
}
