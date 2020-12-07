package gotcha

import (
	"context"
	"unsafe"

	"github.com/1pkg/golocal"
)

// Tracer defines function type that will be traced by gotcha
// it accepts gotcha context that will be seamlessly tracking
// all allocations inside tracer function.
type Tracer func(Context)

// Trace starts memory tracing for provided tracer function.
// Note that trace function could be cobined with each other
// by providing gotcha context to child trace function.
func Trace(ctx context.Context, t Tracer, opts ...ContextOpt) {
	gctx := NewContext(ctx, opts...)
	ptr := uintptr(unsafe.Pointer(gctx.(*gotchactx)))
	golocal.LStore().Set(ptr)
	defer golocal.LStore().Del()
	t(gctx)
}
