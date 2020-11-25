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

// chantype from `runtime.chantype`
type chantype struct {
	tp tp
}

// slice from `runtime.slice`
type slice struct {
	_ unsafe.Pointer
	_ int
	_ int
}

// hchan from `runtime.chan`
type hchan struct{}

// tmpBuf from `runtime.tmpBuf`
type tmpBuf [32]byte

//go:linkname newobject runtime.newobject
func newobject(tp *tp) unsafe.Pointer

//go:linkname reflectUnsafeNew reflect.unsafe_New
func reflectUnsafeNew(tp *tp) unsafe.Pointer

//go:linkname reflectliteUnsafeNew internal/reflectlite.unsafe_New
func reflectliteUnsafeNew(tp *tp) unsafe.Pointer

//go:linkname newarray runtime.newarray
func newarray(typ *tp, n int) unsafe.Pointer

//go:linkname mallocgc runtime.mallocgc
func mallocgc(size uintptr, tp *tp, needzero bool) unsafe.Pointer

//go:linkname makeslice runtime.makeslice
func makeslice(tp *tp, len, cap int) unsafe.Pointer

//go:linkname makeslicecopy runtime.makeslicecopy
func makeslicecopy(tp *tp, tolen int, fromlen int, from unsafe.Pointer) unsafe.Pointer

//go:linkname growslice runtime.growslice
func growslice(tp *tp, old slice, cap int) slice

//go:linkname makechan runtime.makechan
func makechan(tp *chantype, size int) *hchan

//go:linkname rawstring runtime.rawstring
func rawstring(size int) (string, []byte)

//go:linkname rawbyteslice runtime.rawbyteslice
func rawbyteslice(size int) []byte

//go:linkname rawruneslice runtime.rawruneslice
func rawruneslice(size int) []rune

//go:linkname sliceByteToString runtime.slicebytetostring
func sliceByteToString(buf *tmpBuf, ptr *byte, n int) string

// mulUintptr copied from `internal/math.MulUintptr`
func mulUintptr(a, b uintptr) (uintptr, bool) {
	if a|b < 1<<(4*unsafe.Sizeof(uintptr(0))) || a == 0 {
		return a * b, false
	}
	overflow := b > ^uintptr(0)/a
	return a * b, overflow
}

func init() {
	// mallocgc directly
	monkey.Patch(newobject, func(tp *tp) unsafe.Pointer {
		trackAlloc(int(tp.size), 1)
		return mallocgc(tp.size, tp, true)
	})
	monkey.Patch(reflectUnsafeNew, func(tp *tp) unsafe.Pointer {
		trackAlloc(int(tp.size), 1)
		return mallocgc(tp.size, tp, true)
	})
	monkey.Patch(reflectliteUnsafeNew, func(tp *tp) unsafe.Pointer {
		trackAlloc(int(tp.size), 1)
		return mallocgc(tp.size, tp, true)
	})
	monkey.Patch(newarray, func(tp *tp, n int) unsafe.Pointer {
		trackAlloc(int(tp.size), n)
		if n == 1 {
			return mallocgc(tp.size, tp, true)
		}
		mem, overflow := mulUintptr(tp.size, uintptr(n))
		// note: check for `mem > maxAlloc` is omitted here
		if overflow || n < 0 {
			panic("runtime: allocation size out of range")
		}
		return mallocgc(mem, tp, true)
	})
	// slice allocs
	var gmakeSlice, gmakeSliceCopy, ggrowSlice *monkey.PatchGuard
	gmakeSlice = monkey.Patch(makeslice, func(tp *tp, len, cap int) unsafe.Pointer {
		gmakeSlice.Unpatch()
		defer gmakeSlice.Restore()
		trackAlloc(int(tp.size), cap)
		return makeslice(tp, len, cap)
	})
	gmakeSliceCopy = monkey.Patch(makeslicecopy, func(tp *tp, tolen int, fromlen int, from unsafe.Pointer) unsafe.Pointer {
		gmakeSliceCopy.Unpatch()
		defer gmakeSliceCopy.Restore()
		trackAlloc(int(tp.size), tolen)
		return makeslicecopy(tp, tolen, fromlen, from)
	})
	ggrowSlice = monkey.Patch(growslice, func(tp *tp, old slice, cap int) slice {
		ggrowSlice.Unpatch()
		defer ggrowSlice.Restore()
		trackAlloc(int(tp.size), cap)
		return growslice(tp, old, cap)
	})
	// chan allocs
	var gmakeChan *monkey.PatchGuard
	gmakeChan = monkey.Patch(makechan, func(tp *chantype, size int) *hchan {
		gmakeChan.Unpatch()
		defer gmakeChan.Restore()
		trackAlloc(int(tp.tp.size), size)
		return makechan(tp, size)
	})
	// string allocs
	// note that we don't patch `runtime.gobytes`
	// as it seems it's only used by go compiler itself
	var grawString, grawBytes, grawRunes, gSliceBytesString *monkey.PatchGuard
	grawString = monkey.Patch(rawstring, func(size int) (string, []byte) {
		grawString.Unpatch()
		defer grawString.Restore()
		trackAlloc(1, size)
		return rawstring(size)
	})
	grawBytes = monkey.Patch(rawbyteslice, func(size int) []byte {
		grawBytes.Unpatch()
		defer grawBytes.Restore()
		trackAlloc(1, size)
		return rawbyteslice(size)
	})
	grawRunes = monkey.Patch(rawruneslice, func(size int) []rune {
		grawRunes.Unpatch()
		defer grawRunes.Restore()
		trackAlloc(1, size)
		return rawruneslice(size)
	})
	gSliceBytesString = monkey.Patch(sliceByteToString, func(buf *tmpBuf, ptr *byte, n int) string {
		gSliceBytesString.Unpatch()
		defer gSliceBytesString.Restore()
		trackAlloc(1, n)
		return sliceByteToString(buf, ptr, n)
	})
}

func main() {
	var mp map[string]int
	Gotcha(context.Background(), func(ctx Context) {
		mp = make(map[string]int, 100)
		print(ctx.String())
	})
	print(mp)
}
