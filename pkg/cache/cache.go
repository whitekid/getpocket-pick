package cache

import (
	"context"
	"errors"
)

type Interface interface {
	Set(ctx context.Context, key string, value []byte, opts ...setOption) error

	// return ErrNotExists if key not exists
	Get(ctx context.Context, key string) ([]byte, error)

	// return true if key exists
	Has(ctx context.Context, key string) bool
}

var (
	ErrNotExists = errors.New("not exists")
)
