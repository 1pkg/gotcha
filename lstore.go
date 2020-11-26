package gotcha

import (
	"context"

	"github.com/modern-go/gls"
)

type lskey string

const glskey lskey = "glskey"

type Tracer func(Context)

func Trace(ctx context.Context, t Tracer, opts ...ContextOpt) {
	gls.WithGls(func() {
		ctx := NewContext(ctx, opts...)
		gls.Set(glskey, ctx)
		t(ctx)
	})()
}

func trackAlloc(bytes, objects int) {
	if v := gls.Get(glskey); v != nil {
		if ctx, ok := v.(Tracker); ok {
			ctx.Add(int64(bytes), int64(objects), 1)
		}
	}
}
