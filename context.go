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
	Add(bytes, objects, calls uint64)
	Get() (bytes, objects, calls uint64)
	Limit() (lbytes, lobjects, lcalls uint64)
	Check() bool
	Reset()
}

type Context interface {
	context.Context
	String() string
	Tracker
}

type gotchactx struct {
	parent                   context.Context
	bytes, objects, calls    uint64
	lbytes, lobjects, lcalls uint64
}

func NewContext(parent context.Context, opts ...ContextOpt) Context {
	ctx := &gotchactx{parent: parent}
	opts = append([]ContextOpt{
		ContextWithLimitBytes(16 * MiB),
		ContextWithLimitObjects(Mega),
		ContextWithLimitCalls(Kilo),
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
			if ctx.Check() {
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
	if ctx.Check() {
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
		atomic.LoadUint64(&ctx.objects),
		atomic.LoadUint64(&ctx.bytes),
		atomic.LoadUint64(&ctx.calls),
	)
}

func (ctx *gotchactx) Add(bytes, objects, calls uint64) {
	atomic.AddUint64(&ctx.bytes, bytes*objects)
	atomic.AddUint64(&ctx.objects, objects)
	atomic.AddUint64(&ctx.calls, calls)
}

func (ctx *gotchactx) Get() (bytes, objects, calls uint64) {
	return atomic.LoadUint64(&ctx.bytes),
		atomic.LoadUint64(&ctx.objects),
		atomic.LoadUint64(&ctx.calls)
}

func (ctx *gotchactx) Limit() (lbytes, lobjects, lcalls uint64) {
	return atomic.LoadUint64(&ctx.lbytes),
		atomic.LoadUint64(&ctx.lobjects),
		atomic.LoadUint64(&ctx.lcalls)
}

func (ctx *gotchactx) Check() bool {
	return atomic.LoadUint64(&ctx.bytes) > atomic.LoadUint64(&ctx.lbytes) ||
		atomic.LoadUint64(&ctx.bytes) > atomic.LoadUint64(&ctx.lobjects) ||
		atomic.LoadUint64(&ctx.calls) > atomic.LoadUint64(&ctx.lcalls)
}

func (ctx *gotchactx) Reset() {
	atomic.StoreUint64(&ctx.bytes, 0)
	atomic.StoreUint64(&ctx.objects, 0)
	atomic.StoreUint64(&ctx.calls, 0)
}
