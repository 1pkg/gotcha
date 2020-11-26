package gotcha

import (
	"context"
	"sync"

	"github.com/modern-go/gls"
)

type lskey string

const glskey lskey = "glskey"

type Tracer func(Context)

func Trace(ctx context.Context, wg *sync.WaitGroup, t Tracer, opts ...ContextOpt) {
	if wg != nil {
		wg.Add(1)
	}
	go gls.WithGls(func() {
		ctx := NewContext(ctx, opts...)
		gls.Set(glskey, ctx)
		t(ctx)
		if wg != nil {
			wg.Done()
		}
	})()
}

func trackAlloc(bytes, objects int) {
}
