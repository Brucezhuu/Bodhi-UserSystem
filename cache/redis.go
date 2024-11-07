package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

var ctx = context.Background()
var RedisClient *redis.Client

// InitializeRedis 初始化 Redis 客户端
func InitializeRedis(addr, password string, db int) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

// SetCache 设置缓存数据
func SetCache(key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

// GetCache 获取缓存数据
func GetCache(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}
