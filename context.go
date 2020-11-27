package gotcha

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type ContextLimitsExceeded struct {
	Context Context
}

func (err ContextLimitsExceeded) Error() string {
	return fmt.Sprintf("context limits have been exceeded %q", err.Context)
}

type Tracker interface {
	Add(bytes, objects, calls int64)
	Used() (bytes, objects, calls int64)
	Limits() (lbytes, lobjects, lcalls int64)
	Remains() (rbytes, robjects, rcalls int64)
	Exceeded() bool
	Reset()
}

type Context interface {
	context.Context
	String() string
	Tracker
}

type gotchactx struct {
	parent                   context.Context
	bytes, objects, calls    int64
	lbytes, lobjects, lcalls int64
}

func NewContext(parent context.Context, opts ...ContextOpt) Context {
	ctx := &gotchactx{parent: parent}
	opts = append([]ContextOpt{
		ContextWithLimitBytes(64 * MiB),
		ContextWithLimitObjects(Infinity),
		ContextWithLimitCalls(Infinity),
	}, opts...)
	for _, opt := range opts {
		opt(ctx)
	}
	return ctx
}

func (ctx *gotchactx) Deadline() (time.Time, bool) {
	return ctx.parent.Deadline()
}

func (ctx *gotchactx) Done() <-chan struct{} {
	ch := make(chan struct{})
	if ctx.Exceeded() {
		close(ch)
		return ch
	}
	select {
	case <-ctx.parent.Done():
		close(ch)
		return ch
	default:
	}
	// parent context pooling is the simplest solution here
	t := time.NewTicker(time.Millisecond)
	go func() {
		defer t.Stop()
		defer close(ch)
		for {
			select {
			case <-ctx.parent.Done():
				return
			case <-t.C:
				if ctx.Exceeded() {
					return
				}
			}
		}
	}()
	return ch
}

func (ctx *gotchactx) Err() error {
	if err := ctx.parent.Err(); err != nil {
		return err
	}
	if ctx.Exceeded() {
		return ContextLimitsExceeded{Context: ctx}
	}
	return nil
}

func (ctx *gotchactx) Value(key interface{}) interface{} {
	return ctx.parent.Value(key)
}

func (ctx *gotchactx) String() string {
	return fmt.Sprintf(
		"on this context: %d objects has been allocated with total size of %d bytes within %d calls",
		atomic.LoadInt64(&ctx.objects),
		atomic.LoadInt64(&ctx.bytes),
		atomic.LoadInt64(&ctx.calls),
	)
}

func (ctx *gotchactx) Add(bytes, objects, calls int64) {
	atomic.AddInt64(&ctx.bytes, bytes*objects)
	atomic.AddInt64(&ctx.objects, objects)
	atomic.AddInt64(&ctx.calls, calls)
	if pctx, ok := ctx.parent.(Tracker); ok {
		pctx.Add(bytes, objects, calls)
	}
}

func (ctx *gotchactx) Used() (bytes, objects, calls int64) {
	return atomic.LoadInt64(&ctx.bytes),
		atomic.LoadInt64(&ctx.objects),
		atomic.LoadInt64(&ctx.calls)
}

func (ctx *gotchactx) Limits() (lbytes, lobjects, lcalls int64) {
	return atomic.LoadInt64(&ctx.lbytes),
		atomic.LoadInt64(&ctx.lobjects),
		atomic.LoadInt64(&ctx.lcalls)
}

func (ctx *gotchactx) Remains() (rbytes, robjects, rcalls int64) {
	// calculate bytes remains
	bytes := atomic.LoadInt64(&ctx.bytes)
	lbytes := atomic.LoadInt64(&ctx.lbytes)
	switch {
	case lbytes <= Infinity:
		rbytes = Infinity
		if pctx, ok := ctx.parent.(Tracker); ok {
			rbytes, _, _ = pctx.Remains()
		}
	case lbytes > bytes:
		rbytes = lbytes - bytes
	default:
		rbytes = 0
	}
	// calculate objects remains
	objects := atomic.LoadInt64(&ctx.objects)
	lobjects := atomic.LoadInt64(&ctx.lobjects)
	switch {
	case lobjects <= Infinity:
		robjects = Infinity
		if pctx, ok := ctx.parent.(Tracker); ok {
			_, robjects, _ = pctx.Remains()
		}
	case lobjects > objects:
		robjects = lobjects - objects
	default:
		robjects = 0
	}
	// calculate calls remains
	calls := atomic.LoadInt64(&ctx.calls)
	lcalls := atomic.LoadInt64(&ctx.lcalls)
	switch {
	case lcalls <= Infinity:
		rcalls = Infinity
		if pctx, ok := ctx.parent.(Tracker); ok {
			_, _, rcalls = pctx.Remains()
		}
	case lcalls > calls:
		rcalls = lcalls - calls
	default:
		rcalls = 0
	}
	return
}

func (ctx *gotchactx) Exceeded() bool {
	if l := atomic.LoadInt64(&ctx.lbytes); l > Infinity && l < atomic.LoadInt64(&ctx.bytes) {
		return true
	}
	if l := atomic.LoadInt64(&ctx.lobjects); l > Infinity && l < atomic.LoadInt64(&ctx.objects) {
		return true
	}
	if l := atomic.LoadInt64(&ctx.lcalls); l > Infinity && l < atomic.LoadInt64(&ctx.calls) {
		return true
	}
	if pctx, ok := ctx.parent.(Tracker); ok {
		return pctx.Exceeded()
	}
	return false
}

func (ctx *gotchactx) Reset() {
	atomic.StoreInt64(&ctx.bytes, 0)
	atomic.StoreInt64(&ctx.objects, 0)
	atomic.StoreInt64(&ctx.calls, 0)
}
