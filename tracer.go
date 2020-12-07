package gotcha

import "context"

// Tracer defines function type that will be traced by gotcha
// it accepts gotcha context that will be seamlessly tracking
// all allocations inside tracer function.
type Tracer func(Context)

// Trace starts memory tracing for provided tracer function.
// Note that trace function could be cobined with each other
// by providing gotcha context to child trace function.
func Trace(ctx context.Context, t Tracer, opts ...ContextOpt) {
	gctx := NewContext(ctx, opts...)
	glstore.set(gctx)
	defer glstore.del()
	t(gctx)
}
