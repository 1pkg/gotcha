package gotcha

import (
	"context"

	"github.com/modern-go/gls"
)

// lskey is not exported key type for gls
// to not interfire with other gls keys.
type lskey string

// glskey is global lskey gls value.
const glskey lskey = "glskey"

// Tracer defines function type that will be traced by gotcha
// it accepts gotcha context that will be seamlessly tracking
// all allocations inside tracer function.
type Tracer func(Context)

// Trace starts memory tracing for provided tracer function.
// Note that trace function could be cobined with each other
// by providing gotcha context to child trace function.
func Trace(ctx context.Context, t Tracer, opts ...ContextOpt) {
	gls.WithGls(func() {
		ctx := NewContext(ctx, opts...)
		gls.Set(glskey, ctx)
		t(ctx)
	})()
}

// trackAlloc defines global entrypoint to track
// allocations for caller tracer goroutine.
func trackAlloc(bytes, objects int) {
	if v := gls.Get(glskey); v != nil {
		if ctx, ok := v.(Tracker); ok {
			ctx.Add(int64(bytes), int64(objects), 1)
		}
	}
}
