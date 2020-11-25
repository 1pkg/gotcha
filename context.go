package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type Context interface {
	context.Context
	String() string
	Add(bytes, objects uint64)
	Set(lbytes, lobjects, lcalls uint64)
	Get() (bytes, objects, calls uint64)
	Limits() (lbytes, lobjects, lcalls uint64)
	Check() bool
}

type gotchaContext struct {
	parent                   context.Context
	signal                   chan struct{}
	bytes, objects, calls    uint64
	lbytes, lobjects, lcalls uint64
}

func NewContext(ctx context.Context, lbytes, lobjects, lcalls uint64) Context {
	return &gotchaContext{
		parent:   ctx,
		signal:   make(chan struct{}),
		lbytes:   lbytes,
		lobjects: lobjects,
		lcalls:   lcalls,
	}
}

func (ctx *gotchaContext) Deadline() (time.Time, bool) {
	return ctx.parent.Deadline()
}

func (ctx *gotchaContext) Done() <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.parent.Done():
			close(ch)
		case <-ctx.signal:
			close(ch)
		}
	}()
	return ch
}

func (ctx *gotchaContext) Err() error {
	if err := ctx.parent.Err(); err != nil {
		return err
	}
	return nil
}

func (ctx *gotchaContext) Value(key interface{}) interface{} {
	return ctx.parent.Value(key)
}

func (ctx gotchaContext) String() string {
	return fmt.Sprintf(
		"on this context: %d objects has been allocated with total size of %d bytes within %d calls",
		atomic.LoadUint64(&ctx.objects),
		atomic.LoadUint64(&ctx.bytes),
		atomic.LoadUint64(&ctx.calls),
	)
}

func (ctx *gotchaContext) Add(bytes, objects uint64) {
	atomic.StoreUint64(&ctx.bytes, bytes*objects)
	atomic.StoreUint64(&ctx.objects, objects)
	atomic.AddUint64(&ctx.calls, 1)
	if ctx.Check() {
		select {
		case <-ctx.signal:
			return
		default:
			close(ctx.signal)
		}
	}
}

func (ctx *gotchaContext) Set(lbytes, lobjects, lcalls uint64) {
	atomic.StoreUint64(&ctx.lbytes, lbytes)
	atomic.StoreUint64(&ctx.lobjects, lobjects)
	atomic.StoreUint64(&ctx.lcalls, lcalls)
}

func (ctx *gotchaContext) Get() (bytes, objects, calls uint64) {
	return atomic.LoadUint64(&ctx.bytes),
		atomic.LoadUint64(&ctx.objects),
		atomic.LoadUint64(&ctx.calls)
}

func (ctx *gotchaContext) Limits() (lbytes, lobjects, lcalls uint64) {
	return atomic.LoadUint64(&ctx.lbytes),
		atomic.LoadUint64(&ctx.lobjects),
		atomic.LoadUint64(&ctx.lcalls)
}

func (ctx *gotchaContext) Check() bool {
	return atomic.LoadUint64(&ctx.bytes) > atomic.LoadUint64(&ctx.lbytes) ||
		atomic.LoadUint64(&ctx.bytes) > atomic.LoadUint64(&ctx.lobjects) ||
		atomic.LoadUint64(&ctx.calls) > atomic.LoadUint64(&ctx.lcalls)
}
