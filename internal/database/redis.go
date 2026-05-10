package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func BlacklistToken(ctx context.Context, rdb *redis.Client, token string, expiry time.Duration) error {
	return rdb.Set(ctx, "blacklisted:"+token, "1", expiry).Err()
}

func IsTokenBlacklisted(ctx context.Context, rdb *redis.Client, token string) bool {
	result, err := rdb.Exists(ctx, "blacklisted:"+token).Result()
	if err != nil {
		return false
	}
	return result > 0
}
