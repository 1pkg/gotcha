package gotcha

import (
	"context"

	"github.com/modern-go/gls"
)

type lskey string

const glskey lskey = "glskey"

type Tracer func(Context)

func Gotcha(ctx context.Context, gt Tracer, opts ...ContextOpt) {
	gls.WithGls(func() {
		ctx := NewContext(context.Background(), opts...)
		gls.Set(glskey, ctx)
		gt(ctx)
	})()
}

func trackAlloc(bytes, objects int) {
	if v := gls.Get(glskey); v != nil {
		if ctx, ok := v.(Context); ok {
			ctx.Add(uint64(bytes), uint64(objects))
		}
	}
}
