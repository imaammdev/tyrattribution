package redis

import (
	"context"
)

type Client interface {
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, seconds int) error
	Get(ctx context.Context, key string) (string, error)
}