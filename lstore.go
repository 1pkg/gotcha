package main

import (
	"context"

	"github.com/modern-go/gls"
)

type lskey string

const glskey lskey = "glskey"

func trackAlloc(bytes, objects uint64) {
	if v := gls.Get(glskey); v != nil {
		if ctx, ok := v.(Context); ok {
			ctx.Add(bytes, objects)
		}
	}
}

type Tracer func(Context)

func Gotcha(ctx context.Context, gt Tracer) {
	gls.WithGls(func() {
		ctx := NewContext(context.Background(), 100, 10, 10)
		gls.Set(glskey, ctx)
		gt(ctx)
	})()
}
