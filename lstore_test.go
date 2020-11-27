package gotcha

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// epsAlloc defines alloc error delta for test
// as it's tough to carefully track all possible small
// underlying allocations - defining relative deviation buffer simplifies it
const epsAlloc = 0.66

func TestTraceTypes(t *testing.T) {
	t.Run("track no object alloc", func(t *testing.T) {
		Trace(context.Background(), func(ctx Context) {
			for i := 0; i < 1000; i++ {
				_ = i - 1
			}
			b, o, c := ctx.Used()
			require.Equal(t, int64(0), b)
			require.Equal(t, int64(0), o)
			require.Equal(t, int64(0), c)
		})
	})
	t.Run("track new object alloc", func(t *testing.T) {
		type sobj struct {
			a, b int64
		}
		Trace(context.Background(), func(ctx Context) {
			var v *sobj
			for i := 0; i < 1000; i++ {
				v = new(sobj)
			}
			v.a = 0
			v.b = 0
			b, o, c := ctx.Used()
			require.InEpsilon(t, int64(16000), b, epsAlloc)
			require.InEpsilon(t, int64(1000), o, epsAlloc)
			require.GreaterOrEqual(t, c, int64(1000))
		})
	})
	t.Run("track new object alloc &", func(t *testing.T) {
		type sobj struct {
			a, b int64
		}
		Trace(context.Background(), func(ctx Context) {
			var v *sobj
			for i := 0; i < 1000; i++ {
				v = &sobj{a: 1, b: 1}
			}
			v.a = 0
			v.b = 0
			b, o, c := ctx.Used()
			require.InEpsilon(t, int64(16000), b, epsAlloc)
			require.InEpsilon(t, int64(1000), o, epsAlloc)
			require.GreaterOrEqual(t, c, int64(1000))
		})
	})
	t.Run("track new object alloc reflect", func(t *testing.T) {
		type sobj struct {
			a, b int64
		}
		Trace(context.Background(), func(ctx Context) {
			var v *sobj = &sobj{}
			tp := reflect.ValueOf(*v).Type()
			for i := 0; i < 100; i++ {
				vref := reflect.New(tp)
				v = vref.Interface().(*sobj)
			}
			v.a = 0
			v.b = 0
			b, o, c := ctx.Used()
			require.InEpsilon(t, int64(1600), b, epsAlloc)
			require.InEpsilon(t, int64(100), o, epsAlloc)
			require.GreaterOrEqual(t, c, int64(100))
		})
	})
	t.Run("track make slice alloc", func(t *testing.T) {
		Trace(context.Background(), func(ctx Context) {
			var v []int64
			for i := 0; i < 1000; i++ {
				v = make([]int64, 1, 10)
			}
			v[0] = 0
			b, o, c := ctx.Used()
			require.InEpsilon(t, int64(80000), b, epsAlloc)
			require.InEpsilon(t, int64(10000), o, epsAlloc)
			require.GreaterOrEqual(t, c, int64(1000))
		})
	})
	t.Run("track make slice copy alloc", func(t *testing.T) {
		Trace(context.Background(), func(ctx Context) {
			var v []int64
			vc := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			for i := 0; i < 1000; i++ {
				v = make([]int64, 10)
				copy(v, vc)
			}
			v[0] = 0
			b, o, c := ctx.Used()
			require.InEpsilon(t, int64(80000), b, epsAlloc)
			require.InEpsilon(t, int64(10000), o, epsAlloc)
			require.GreaterOrEqual(t, c, int64(1000))
		})
	})
	t.Run("track make slice append alloc", func(t *testing.T) {
		Trace(context.Background(), func(ctx Context) {
			var v []int64
			for i := 0; i < 1000; i++ {
				v = make([]int64, 10)
				v = append(v, 1)
			}
			v[0] = 0
			b, o, c := ctx.Used()
			require.InEpsilon(t, int64(160000), b, epsAlloc)
			require.InEpsilon(t, int64(20000), o, epsAlloc)
			require.GreaterOrEqual(t, c, int64(2000))
		})
	})
	t.Run("track make map alloc", func(t *testing.T) {
		Trace(context.Background(), func(ctx Context) {
			var v map[string]int32
			for i := 0; i < 100; i++ {
				v = make(map[string]int32, 15)
			}
			v[""] = 0
			b, o, c := ctx.Used()
			require.GreaterOrEqual(t, b, int64(6000))
			require.GreaterOrEqual(t, o, int64(100))
			require.GreaterOrEqual(t, c, int64(100))
		})
	})
	t.Run("track make chan alloc", func(t *testing.T) {
		Trace(context.Background(), func(ctx Context) {
			var v chan [12]byte
			for i := 0; i < 100; i++ {
				v = make(chan [12]byte, 10)
			}
			close(v)
			b, o, c := ctx.Used()
			require.GreaterOrEqual(t, b, int64(1200))
			require.GreaterOrEqual(t, o, int64(1000))
			require.GreaterOrEqual(t, c, int64(100))
		})
	})
	t.Run("track string bytes alloc", func(t *testing.T) {
		Trace(context.Background(), func(ctx Context) {
			cs := "foo | | bar"
			var v []byte
			for i := 0; i < 1000; i++ {
				v = []byte(cs)
			}
			_ = len(v)
			b, o, c := ctx.Used()
			require.InEpsilon(t, int64(22000), b, epsAlloc)
			require.InEpsilon(t, int64(11000), o, epsAlloc)
			require.GreaterOrEqual(t, c, int64(1000))
		})
	})
	t.Run("track string runes alloc", func(t *testing.T) {
		Trace(context.Background(), func(ctx Context) {
			cs := "foo | | bar"
			var v []rune
			for i := 0; i < 1000; i++ {
				v = []rune(cs)
			}
			_ = len(v)
			b, o, c := ctx.Used()
			require.InEpsilon(t, int64(22000), b, epsAlloc)
			require.InEpsilon(t, int64(11000), o, epsAlloc)
			require.GreaterOrEqual(t, c, int64(1000))
		})
	})
	t.Run("track bytes string alloc", func(t *testing.T) {
		Trace(context.Background(), func(ctx Context) {
			cb := []byte("foo | | bar")
			var v string
			for i := 0; i < 1000; i++ {
				v = string(cb)
			}
			_ = len(v)
			b, o, c := ctx.Used()
			require.InEpsilon(t, int64(22000), b, epsAlloc)
			require.InEpsilon(t, int64(11000), o, epsAlloc)
			require.GreaterOrEqual(t, c, int64(1000))
		})
	})
	t.Run("track new object alloc complex", func(t *testing.T) {
		type cobj struct {
			mp   map[string][]string
			self *cobj
		}
		Trace(context.Background(), func(ctx Context) {
			var v *cobj
			for i := 0; i < 100; i++ {
				v = &cobj{
					mp: map[string][]string{
						"foo": []string{"foo | | bar"},
					},
					self: &cobj{},
				}
			}
			_ = v.self
			b, o, c := ctx.Used()
			require.GreaterOrEqual(t, b, int64(1400))
			require.GreaterOrEqual(t, o, int64(100))
			require.GreaterOrEqual(t, c, int64(100))
		})
	})
}

func TestTraceHierarchy(t *testing.T) {
	Trace(context.Background(), func(ctx Context) {
		var v1, v2, v3 []int64
		Trace(ctx, func(ctx Context) {
			v1 = make([]int64, 5, 10)
			b, o, c := ctx.Used()
			require.GreaterOrEqual(t, b, int64(80))
			require.GreaterOrEqual(t, o, int64(10))
			require.GreaterOrEqual(t, c, int64(1))
		})
		Trace(ctx, func(ctx Context) {
			v2 = make([]int64, 5, 20)
			b, o, c := ctx.Used()
			require.GreaterOrEqual(t, b, int64(160))
			require.GreaterOrEqual(t, o, int64(20))
			require.GreaterOrEqual(t, c, int64(1))
			Trace(ctx, func(ctx Context) {
				v3 = make([]int64, 5, 10)
				b, o, c := ctx.Used()
				require.GreaterOrEqual(t, b, int64(80))
				require.GreaterOrEqual(t, o, int64(10))
				require.GreaterOrEqual(t, c, int64(1))
			})
			Trace(ctx, func(ctx Context) {
				v2 = make([]int64, 5, 20)
				b, o, c := ctx.Used()
				require.GreaterOrEqual(t, b, int64(160))
				require.GreaterOrEqual(t, o, int64(20))
				require.GreaterOrEqual(t, c, int64(1))
			})
			var wg sync.WaitGroup
			wg.Add(1)
			go Trace(ctx, func(ctx Context) {
				v2 = make([]int64, 5, 20)
				b, o, c := ctx.Used()
				require.GreaterOrEqual(t, b, int64(160))
				require.GreaterOrEqual(t, o, int64(20))
				require.GreaterOrEqual(t, c, int64(1))
				wg.Done()
			})
			wg.Wait()
		})
		v1[0] = 0
		v2[0] = 0
		v3[0] = 0
		b, o, c := ctx.Used()
		require.GreaterOrEqual(t, b, int64(640))
		require.GreaterOrEqual(t, o, int64(80))
		require.GreaterOrEqual(t, c, int64(5))
	})
}
