package gotcha

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTracker(t *testing.T) {
	t.Run("context with positive limits", func(t *testing.T) {
		ctx := NewContext(
			context.Background(),
			ContextWithLimitBytes(10),
			ContextWithLimitObjects(5),
			ContextWithLimitCalls(3),
		)
		b, o, c := ctx.Used()
		require.Equal(t, int64(0), b)
		require.Equal(t, int64(0), o)
		require.Equal(t, int64(0), c)
		lb, lo, lc := ctx.Limits()
		require.Equal(t, int64(10), lb)
		require.Equal(t, int64(5), lo)
		require.Equal(t, int64(3), lc)
		rb, ro, rc := ctx.Remains()
		require.Equal(t, int64(10), rb)
		require.Equal(t, int64(5), ro)
		require.Equal(t, int64(3), rc)
		require.False(t, ctx.Exceeded())
		ctx.Add(2, 4, 2)
		b, o, c = ctx.Used()
		require.Equal(t, int64(8), b)
		require.Equal(t, int64(4), o)
		require.Equal(t, int64(2), c)
		lb, lo, lc = ctx.Limits()
		require.Equal(t, int64(10), lb)
		require.Equal(t, int64(5), lo)
		require.Equal(t, int64(3), lc)
		rb, ro, rc = ctx.Remains()
		require.Equal(t, int64(2), rb)
		require.Equal(t, int64(1), ro)
		require.Equal(t, int64(1), rc)
		require.False(t, ctx.Exceeded())
		ctx.Add(3, 1, 1)
		b, o, c = ctx.Used()
		require.Equal(t, int64(11), b)
		require.Equal(t, int64(5), o)
		require.Equal(t, int64(3), c)
		lb, lo, lc = ctx.Limits()
		require.Equal(t, int64(10), lb)
		require.Equal(t, int64(5), lo)
		require.Equal(t, int64(3), lc)
		rb, ro, rc = ctx.Remains()
		require.Equal(t, int64(0), rb)
		require.Equal(t, int64(0), ro)
		require.Equal(t, int64(0), rc)
		require.True(t, ctx.Exceeded())
		ctx.Reset()
		b, o, c = ctx.Used()
		require.Equal(t, int64(0), b)
		require.Equal(t, int64(0), o)
		require.Equal(t, int64(0), c)
		lb, lo, lc = ctx.Limits()
		require.Equal(t, int64(10), lb)
		require.Equal(t, int64(5), lo)
		require.Equal(t, int64(3), lc)
		rb, ro, rc = ctx.Remains()
		require.Equal(t, int64(10), rb)
		require.Equal(t, int64(5), ro)
		require.Equal(t, int64(3), rc)
		require.False(t, ctx.Exceeded())
		ctx.Add(0, 0, 10)
		b, o, c = ctx.Used()
		require.Equal(t, int64(0), b)
		require.Equal(t, int64(0), o)
		require.Equal(t, int64(10), c)
		lb, lo, lc = ctx.Limits()
		require.Equal(t, int64(10), lb)
		require.Equal(t, int64(5), lo)
		require.Equal(t, int64(3), lc)
		rb, ro, rc = ctx.Remains()
		require.Equal(t, int64(10), rb)
		require.Equal(t, int64(5), ro)
		require.Equal(t, int64(0), rc)
		require.True(t, ctx.Exceeded())
	})
	t.Run("context with infinity limits", func(t *testing.T) {
		ctx := NewContext(
			context.Background(),
			ContextWithLimitBytes(5*Infinity),
			ContextWithLimitObjects(3),
		)
		b, o, c := ctx.Used()
		require.Equal(t, int64(0), b)
		require.Equal(t, int64(0), o)
		require.Equal(t, int64(0), c)
		lb, lo, lc := ctx.Limits()
		require.Equal(t, int64(5*Infinity), lb)
		require.Equal(t, int64(3), lo)
		require.Equal(t, int64(Infinity), lc)
		rb, ro, rc := ctx.Remains()
		require.Equal(t, int64(Infinity), rb)
		require.Equal(t, int64(3), ro)
		require.Equal(t, int64(Infinity), rc)
		require.False(t, ctx.Exceeded())
		ctx.Add(Petabyte, 1, Exa)
		b, o, c = ctx.Used()
		require.Equal(t, int64(Petabyte), b)
		require.Equal(t, int64(1), o)
		require.Equal(t, int64(Exa), c)
		lb, lo, lc = ctx.Limits()
		require.Equal(t, int64(5*Infinity), lb)
		require.Equal(t, int64(3), lo)
		require.Equal(t, int64(Infinity), lc)
		rb, ro, rc = ctx.Remains()
		require.Equal(t, int64(Infinity), rb)
		require.Equal(t, int64(2), ro)
		require.Equal(t, int64(Infinity), rc)
		require.False(t, ctx.Exceeded())
		ctx.Add(Petabyte, 5, Exa)
		b, o, c = ctx.Used()
		require.Equal(t, int64(6*Petabyte), b)
		require.Equal(t, int64(6), o)
		require.Equal(t, int64(2*Exa), c)
		lb, lo, lc = ctx.Limits()
		require.Equal(t, int64(5*Infinity), lb)
		require.Equal(t, int64(3), lo)
		require.Equal(t, int64(Infinity), lc)
		rb, ro, rc = ctx.Remains()
		require.Equal(t, int64(Infinity), rb)
		require.Equal(t, int64(0), ro)
		require.Equal(t, int64(Infinity), rc)
		require.True(t, ctx.Exceeded())
		ctx.Reset()
		b, o, c = ctx.Used()
		require.Equal(t, int64(0), b)
		require.Equal(t, int64(0), o)
		require.Equal(t, int64(0), c)
		lb, lo, lc = ctx.Limits()
		require.Equal(t, int64(5*Infinity), lb)
		require.Equal(t, int64(3), lo)
		require.Equal(t, int64(Infinity), lc)
		rb, ro, rc = ctx.Remains()
		require.Equal(t, int64(Infinity), rb)
		require.Equal(t, int64(3), ro)
		require.Equal(t, int64(Infinity), rc)
		require.False(t, ctx.Exceeded())
	})
	t.Run("context with infinity parent", func(t *testing.T) {
		pctx := NewContext(
			context.Background(),
			ContextWithLimitBytes(10),
			ContextWithLimitObjects(5),
			ContextWithLimitCalls(3),
		)
		ctx := NewContext(
			pctx,
			ContextWithLimitBytes(Infinity),
			ContextWithLimitObjects(Infinity),
			ContextWithLimitCalls(Infinity),
		)
		b, o, c := ctx.Used()
		require.Equal(t, int64(0), b)
		require.Equal(t, int64(0), o)
		require.Equal(t, int64(0), c)
		lb, lo, lc := ctx.Limits()
		require.Equal(t, int64(Infinity), lb)
		require.Equal(t, int64(Infinity), lo)
		require.Equal(t, int64(Infinity), lc)
		rb, ro, rc := ctx.Remains()
		require.Equal(t, int64(10), rb)
		require.Equal(t, int64(5), ro)
		require.Equal(t, int64(3), rc)
		require.False(t, ctx.Exceeded())
		ctx.Add(5, 3, 3)
		b, o, c = ctx.Used()
		require.Equal(t, int64(15), b)
		require.Equal(t, int64(3), o)
		require.Equal(t, int64(3), c)
		lb, lo, lc = ctx.Limits()
		require.Equal(t, int64(Infinity), lb)
		require.Equal(t, int64(Infinity), lo)
		require.Equal(t, int64(Infinity), lc)
		rb, ro, rc = ctx.Remains()
		require.Equal(t, int64(0), rb)
		require.Equal(t, int64(2), ro)
		require.Equal(t, int64(0), rc)
		require.True(t, ctx.Exceeded())
		ctx.Reset()
		b, o, c = ctx.Used()
		require.Equal(t, int64(0), b)
		require.Equal(t, int64(0), o)
		require.Equal(t, int64(0), c)
		lb, lo, lc = ctx.Limits()
		require.Equal(t, int64(Infinity), lb)
		require.Equal(t, int64(Infinity), lo)
		require.Equal(t, int64(Infinity), lc)
		rb, ro, rc = ctx.Remains()
		require.Equal(t, int64(0), rb)
		require.Equal(t, int64(2), ro)
		require.Equal(t, int64(0), rc)
		require.True(t, ctx.Exceeded())
	})
}

func TestContext(t *testing.T) {
	t.Run("context with parent background", func(t *testing.T) {
		ctx := NewContext(
			context.Background(),
			ContextWithLimitBytes(10),
			ContextWithLimitObjects(5),
			ContextWithLimitCalls(3),
		)
		_, dok := ctx.Deadline()
		require.False(t, dok)
		ch := ctx.Done()
		time.Sleep(3 * time.Millisecond)
		select {
		case <-ch:
			require.False(t, true)
		default:
		}
		require.NoError(t, ctx.Err())
		require.Nil(t, ctx.Value("test"))
		ctx.Add(2, 5, 5)
		_, dok = ctx.Deadline()
		require.False(t, dok)
		ch = ctx.Done()
		select {
		case <-ch:
		default:
			require.False(t, true)
		}
		require.Error(t, ctx.Err())
		require.EqualValues(t, ContextLimitsExceeded{Context: ctx}, ctx.Err())
		require.EqualValues(t, `context limits have been exceeded "on this context: 5 objects has been allocated with total size of 10 bytes within 5 calls"`, ctx.Err().Error())
		require.Nil(t, ctx.Value("test"))
		ctx.Reset()
		_, dok = ctx.Deadline()
		require.False(t, dok)
		ch = ctx.Done()
		time.Sleep(3 * time.Millisecond)
		select {
		case <-ch:
			require.False(t, true)
		default:
		}
		require.NoError(t, ctx.Err())
		require.Nil(t, ctx.Value("test"))
		ctx.Add(2, 5, 5)
		time.Sleep(3 * time.Millisecond)
		select {
		case <-ch:
		default:
			require.False(t, true)
		}
	})

	t.Run("context with parent canceled", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(
			//nolint
			context.WithValue(
				context.Background(),
				"test",
				100,
			),
			time.Now().UTC().Add(time.Nanosecond),
		)
		cancel()
		ctx = NewContext(
			ctx,
			ContextWithLimitBytes(10),
			ContextWithLimitObjects(5),
			ContextWithLimitCalls(3),
		)
		_, dok := ctx.Deadline()
		require.True(t, dok)
		ch := ctx.Done()
		time.Sleep(3 * time.Millisecond)
		select {
		case <-ch:
		default:
			require.False(t, true)
		}
		require.Error(t, ctx.Err())
		require.Equal(t, 100, ctx.Value("test"))
	})
}
