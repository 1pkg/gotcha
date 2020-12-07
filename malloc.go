package gotcha

import (
	"C"
	"unsafe"

	"github.com/1pkg/gomonkey"
)

// tp from `runtime._type`
type tp struct {
	size uintptr
}

//go:linkname mallocgc runtime.mallocgc
func mallocgc(size uintptr, tp *tp, needzero bool) unsafe.Pointer

// init patches some existing memory allocation runtime entrypoints
// - direct objects allocation
// - arrays allocation
// - slice allocation
// - map allocation (solved by arrays allocation)
// - chan allocation
// - strings/bytes/runes allocation
// Note that some patches with monkey patch are causing loops, for e.g. `newarray`, `growslice`.
// So for them implementation either slightly changed - newarray or not patched at all - `growslice`.
// For `growslice` it still possible to make the same workaround as done for `newarray`,
// but it will require to copy and support great amount of code from runtime
// which is not correlating with the goal of this project, so `growslice` is skipped for now.
// Note that some function from `interface.conv` family are not patched neither which will cause
// untracked allocation for code like `vt, ok := var.(type)`.
// Note that `runtime.gobytes` is not patched as well as it seems it's only used by go compiler itself.
// Note that only functions from `mallocgc` family are patched, but runtime has much more allocation tricks
// that won't be traced by gotcha, like direct `malloc` sys calls, etc.
func init() {
	glstore = &lstore{
		mp:  make(map[int64]Context, 1024),
		cap: 1024,
	}
	gomonkey.PermanentDecorate(mallocgc, func(size uintptr, tp *tp, needzero bool) unsafe.Pointer {
		if ctx := glstore.get(); ctx != nil {
			// trace allocations for caller tracer goroutine.
			bytes := int64(size)
			objs := int64(1)
			if tp != nil {
				bytes = int64(tp.size)
				objs = int64(size) / bytes
			}
			ctx.Add(bytes, objs, 1)
		}
		return nil
	}, 24, 53, []byte{
		0x48, 0x83, 0xec, 0x28, // sub rsp,0x28
		0x48, 0x8b, 0x44, 0x24, 0x30, // mov rax,QWORD PTR [rsp+0x30]
		0x48, 0x89, 0x04, 0x24, // mov QWORD PTR [rsp],rax
		0x48, 0x8b, 0x44, 0x24, 0x38, // mov rax,QWORD PTR [rsp+0x38]
		0x48, 0x89, 0x44, 0x24, 0x08, // mov QWORD PTR [rsp],rax
	}, []byte{
		0x48, 0x83, 0xc4, 0x28, // add rsp,0x28
		0x48, 0x81, 0xec, 0x98, 0x00, 0x00, 0x00, // sub rsp,0x98
	})
}
