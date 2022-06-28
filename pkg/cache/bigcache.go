package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/allegro/bigcache/v3"
)

func NewBigCache() Interface {
	config := bigcache.DefaultConfig(time.Hour)
	config.CleanWindow = time.Minute
	cache, _ := bigcache.NewBigCache(config)
	return &bigCacheImpl{
		cache: cache,
	}
}

type bigCacheImpl struct {
	cache *bigcache.BigCache
}

var _ Interface = (*bigCacheImpl)(nil)

func (b *bigCacheImpl) Set(ctx context.Context, key string, value []byte, opts ...setOption) error {
	option := applySetOptions(opts)

	if err := b.cache.Set(key, value); err != nil {
		return err
	}

	if option.expire != 0 {
		timez := time.Now().Add(option.expire).Format(time.RFC3339)
		b.cache.Set(fmt.Sprintf("%s/expire", key), []byte(timez))
	}

	return nil
}

func (b *bigCacheImpl) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := b.cache.Get(key)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			return nil, ErrNotExists
		}
		return nil, err
	}

	expireb, err := b.cache.Get(fmt.Sprintf("%s/expire", key))
	if err != nil {
		return data, nil
	}

	expire, err := time.Parse(time.RFC3339, string(expireb))
	if err != nil {
		return data, nil
	}

	if expire.Before(time.Now()) {
		return nil, ErrNotExists
	}

	return data, nil
}

func (b *bigCacheImpl) Has(ctx context.Context, key string) bool {
	_, err := b.cache.Get(key)
	return err == nil
}
