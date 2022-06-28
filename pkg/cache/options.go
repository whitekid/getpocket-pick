package cache

import (
	"time"

	"github.com/whitekid/goxp/fx"
)

func WithExpire(expire time.Duration) setOption {
	return newFuncSetOption(func(o *setOptions) { o.expire = expire })
}

type setOptions struct {
	expire time.Duration
}

type setOption interface {
	apply(o *setOptions)
}

type funcSetOption struct {
	f func(o *setOptions)
}

func (f *funcSetOption) apply(o *setOptions) { f.f(o) }

func newFuncSetOption(f func(o *setOptions)) setOption {
	return &funcSetOption{f: f}
}

func applySetOptions(opts []setOption) *setOptions {
	opt := &setOptions{}
	fx.ForEach(opts, func(_ int, o setOption) { o.apply(opt) })
	return opt
}
