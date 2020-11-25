package main

import (
	"C"
	"context"
	"unsafe"

	"bou.ke/monkey"
)

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

func init() {
	// mallocgc directly
	monkey.Patch(newobject, func(tp *tp) unsafe.Pointer {
		trackAlloc(uint64(tp.size), 1)
		return mallocgc(tp.size, tp, true)
	})
	monkey.Patch(reflectUnsafeNew, func(tp *tp) unsafe.Pointer {
		trackAlloc(uint64(tp.size), 1)
		return mallocgc(tp.size, tp, true)
	})
	monkey.Patch(reflectliteUnsafeNew, func(tp *tp) unsafe.Pointer {
		trackAlloc(uint64(tp.size), 1)
		return mallocgc(tp.size, tp, true)
	})
	// slice allocs
	var gmakeSlice, gmakeSliceCopy, ggrowSlice *monkey.PatchGuard
	gmakeSlice = monkey.Patch(makeslice, func(tp *tp, len, cap int) unsafe.Pointer {
		gmakeSlice.Unpatch()
		defer gmakeSlice.Restore()
		trackAlloc(uint64(tp.size), uint64(cap))
		return makeslice(tp, len, cap)
	})
	gmakeSliceCopy = monkey.Patch(makeslicecopy, func(tp *tp, tolen int, fromlen int, from unsafe.Pointer) unsafe.Pointer {
		gmakeSliceCopy.Unpatch()
		defer gmakeSliceCopy.Restore()
		trackAlloc(uint64(tp.size), uint64(tolen))
		return makeslicecopy(tp, tolen, fromlen, from)
	})
	ggrowSlice = monkey.Patch(growslice, func(tp *tp, old slice, cap int) slice {
		ggrowSlice.Unpatch()
		defer ggrowSlice.Restore()
		trackAlloc(uint64(tp.size), uint64(cap))
		return growslice(tp, old, cap)
	})
}

func main() {
	var a []int
	Gotcha(context.Background(), func(ctx Context) {
		a = make([]int, 100, 105)
		print(ctx.String())
	})
	a[1] = 13
}
