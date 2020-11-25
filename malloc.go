package main

import (
	"C"
	"unsafe"

	"bou.ke/monkey"
	"github.com/modern-go/gls"
)

type lskey string

const glskey lskey = "glskey"

// tp from `runtime._type`
type tp struct {
	size uintptr
}

// slice from `runtime.slice`
type slice struct {
	_ unsafe.Pointer
	_ int
	_ int
}

//go:linkname mallocgc runtime.mallocgc
func mallocgc(size uintptr, tp *tp, needzero bool) unsafe.Pointer

//go:linkname makeslice runtime.makeslice
func makeslice(tp *tp, len, cap int) unsafe.Pointer

//go:linkname makeslicecopy runtime.makeslicecopy
func makeslicecopy(tp *tp, tolen int, fromlen int, from unsafe.Pointer) unsafe.Pointer

//go:linkname growslice runtime.growslice
func growslice(tp *tp, old slice, cap int) slice

func update(bytes, objects int64) {
	if v := gls.Get(glskey); v != nil {
		if ctx, ok := v.(*Context); ok {
			ctx.add(bytes, objects)
			print("DEBUG", bytes, "   ", objects, "\n")
		}
	}
}

func main() {
	var gmakeSlice, gmakeSliceCopy, ggrowSlice *monkey.PatchGuard
	gmakeSlice = monkey.Patch(makeslice, func(tp *tp, len, cap int) unsafe.Pointer {
		gmakeSlice.Unpatch()
		defer gmakeSlice.Restore()
		update(int64(tp.size), int64(cap))
		return makeslice(tp, len, cap)
	})
	gmakeSliceCopy = monkey.Patch(makeslicecopy, func(tp *tp, tolen int, fromlen int, from unsafe.Pointer) unsafe.Pointer {
		gmakeSliceCopy.Unpatch()
		defer gmakeSliceCopy.Restore()
		update(int64(tp.size), int64(tolen))
		return makeslicecopy(tp, tolen, fromlen, from)
	})
	ggrowSlice = monkey.Patch(growslice, func(tp *tp, old slice, cap int) slice {
		ggrowSlice.Unpatch()
		defer ggrowSlice.Restore()
		update(int64(tp.size), int64(cap))
		return growslice(tp, old, cap)
	})
	gls.WithGls(func() {
		ctx := &Context{}
		gls.Set(glskey, ctx)
		a := make([]int, 5, 5)
		for i := 0; i < 10; i++ {
			a = make([]int, 1, 1)
			a = append(a, 1, 2, 3)
		}
		print(ctx.String())
	})()
}
