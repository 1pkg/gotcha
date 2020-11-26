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
	Get() (bytes, objects, calls int64)
	Limit() (lbytes, lobjects, lcalls int64)
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
	t := time.NewTicker(time.Millisecond)
	go func() {
		select {
		case <-ctx.parent.Done():
			close(ch)
		case <-t.C:
			if ctx.Exceeded() {
				close(ch)
				t.Stop()
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
	if ctx, ok := ctx.parent.(Tracker); ok {
		ctx.Add(bytes, objects, calls)
	}
}

func (ctx *gotchactx) Get() (bytes, objects, calls int64) {
	return atomic.LoadInt64(&ctx.bytes),
		atomic.LoadInt64(&ctx.objects),
		atomic.LoadInt64(&ctx.calls)
}

func (ctx *gotchactx) Limit() (lbytes, lobjects, lcalls int64) {
	return atomic.LoadInt64(&ctx.lbytes),
		atomic.LoadInt64(&ctx.lobjects),
		atomic.LoadInt64(&ctx.lcalls)
}

func (ctx *gotchactx) Exceeded() bool {
	if l := atomic.LoadInt64(&ctx.lbytes); l != Infinity && l < atomic.LoadInt64(&ctx.bytes) {
		return true
	}
	if l := atomic.LoadInt64(&ctx.lobjects); l != Infinity && l < atomic.LoadInt64(&ctx.objects) {
		return true
	}
	if l := atomic.LoadInt64(&ctx.lcalls); l != Infinity && l < atomic.LoadInt64(&ctx.calls) {
		return true
	}
	if ctx, ok := ctx.parent.(Tracker); ok {
		return ctx.Exceeded()
	}
	return false
}

func (ctx *gotchactx) Reset() {
	atomic.StoreInt64(&ctx.bytes, 0)
	atomic.StoreInt64(&ctx.objects, 0)
	atomic.StoreInt64(&ctx.calls, 0)
}
