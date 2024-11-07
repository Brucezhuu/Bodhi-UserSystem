package config

import (
	"UserSystem/cache"
	"os"
	"strconv"
)

func InitializeRedisConfig() {
	addr := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	cache.InitializeRedis(addr, password, db)
}
