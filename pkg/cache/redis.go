package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisCacheImpl struct {
	client *redis.Client
}

var _ Interface = (*redisCacheImpl)(nil)

func NewRedis(r *redis.Client) Interface {
	return &redisCacheImpl{
		client: r,
	}
}

func (r *redisCacheImpl) Set(ctx context.Context, key string, value []byte, opts ...setOption) error {
	option := applySetOptions(opts)

	_, err := r.client.Set(ctx, key, value, option.expire).Result()
	return err
}

func (r *redisCacheImpl) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrNotExists
	}

	return result, err
}

func (r *redisCacheImpl) Has(ctx context.Context, key string) bool {
	exists, _ := r.client.Exists(ctx, key).Result()
	return exists != 0
}
