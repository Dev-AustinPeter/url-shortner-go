package cachemanager

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type CacheManager struct {
	rdb *redis.Client
	log zerolog.Logger
}

func NewCacheManager(rdb *redis.Client, log zerolog.Logger) *CacheManager {
	return &CacheManager{
		rdb: rdb,
		log: log,
	}
}

func (cm *CacheManager) Set(ctx context.Context, key string, value string, ttl int) error {

	return cm.rdb.Set(ctx, key, value, time.Duration(ttl*int(time.Minute))).Err()
}

func (cm *CacheManager) Get(ctx context.Context, key string) (string, error) {
	return cm.rdb.Get(ctx, key).Result()
}
