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

//go:linkname newobject runtime.newobject
func newobject(tp *tp) unsafe.Pointer

//go:linkname reflectUnsafeNew reflect.unsafe_New
func reflectUnsafeNew(typ *tp) unsafe.Pointer

//go:linkname reflectliteUnsafeNew internal/reflectlite.unsafe_New
func reflectliteUnsafeNew(typ *tp) unsafe.Pointer

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
			print("DEBUG", "   ", bytes, "   ", objects, "\n")
		}
	}
}

func main() {
	// mallocgc directly
	monkey.Patch(newobject, func(tp *tp) unsafe.Pointer {
		update(int64(tp.size), 1)
		return mallocgc(tp.size, tp, true)
	})
	monkey.Patch(reflectUnsafeNew, func(tp *tp) unsafe.Pointer {
		update(int64(tp.size), 1)
		return mallocgc(tp.size, tp, true)
	})
	monkey.Patch(reflectliteUnsafeNew, func(tp *tp) unsafe.Pointer {
		update(int64(tp.size), 1)
		return mallocgc(tp.size, tp, true)
	})
	// slice allocs
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
		print(ctx.String())
	})()
}
