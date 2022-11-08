package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/zstd"
	"github.com/allegro/bigcache/v3"
	"github.com/pkg/errors"
)

func NewBigCache(ctx context.Context) Interface {
	config := bigcache.DefaultConfig(time.Hour)
	config.CleanWindow = time.Minute
	cache, _ := bigcache.New(ctx, config)
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

	value, err := zstd.Compress(nil, value)
	if err != nil {
		return err
	}

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

	data, err = zstd.Decompress(nil, data)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to decompress")
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
