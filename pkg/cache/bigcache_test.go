package cache

import (
	"context"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := bigcache.DefaultConfig(time.Millisecond * 100)
	config.CleanWindow = time.Second

	cache, _ := bigcache.New(ctx, config)
	require.NoError(t, cache.Set("hello", []byte("world")))
	value, err := cache.Get("hello")
	require.NoError(t, err)
	require.Equal(t, []byte("world"), value)

	// eviction
	{
		time.Sleep(time.Second * 2)
		value, err := cache.Get("hello")
		require.Equal(t, bigcache.ErrEntryNotFound, err)
		require.NotEqual(t, []byte("world"), value)
	}
}
