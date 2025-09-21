package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"tyrattribution/config"

	"github.com/redis/go-redis/v9"
)

type ClientWrapper struct {
	client *redis.Client
}

func NewClient(cfg *config.Config) (Client, error) {
	redisURL := cfg.REDISURL
	redisPassword := cfg.REDISPassword
	redisDBStr := cfg.REDISDBStr

	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		redisDB = 0
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &ClientWrapper{client: client}, nil
}

func (r *ClientWrapper) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *ClientWrapper) Expire(ctx context.Context, key string, seconds int) error {
	return r.client.Expire(ctx, key, time.Duration(seconds)*time.Second).Err()
}

func (r *ClientWrapper) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}
