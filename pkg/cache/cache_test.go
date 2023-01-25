package cache

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestCacher(t *testing.T) {
	s := miniredis.RunT(t)
	r := redis.NewClient(&redis.Options{
		Addr:     s.Addr(),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer s.Close()

	type args struct {
		cacher Interface
	}
	tests := [...]struct {
		name string
		args args
	}{
		{"bigcache", args{NewBigCache(context.Background())}},
		{"redis", args{NewRedis(r)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			key := "sample-key"
			value := []byte("sample-value")

			cacher := tt.args.cacher
			require.False(t, cacher.Has(ctx, key))
			require.NoError(t, cacher.Set(ctx, key, value))

			got, err := cacher.Get(ctx, key)
			require.NoError(t, err)
			require.Equal(t, value, got)

			require.True(t, cacher.Has(ctx, key), "cache shout has value")
		})
	}
}
