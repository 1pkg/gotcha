package gotcha

import (
	"C"
	"os"
	"strconv"
	"unsafe"

	"github.com/1pkg/golocal"
	"github.com/1pkg/gomonkey"
)

// tp from `runtime._type`
type tp struct {
	size uintptr
}

//go:linkname mallocgc runtime.mallocgc
func mallocgc(size uintptr, tp *tp, needzero bool) unsafe.Pointer

// init patches main mallocgc allocation runtime entrypoint
// it also sets goroutine local storage tracers capacity from env var
// note that pacthing will only work on amd64 arch.
func init() {
	// set up local store for malloc
	maxTracers := int64(golocal.DefaultCapacity)
	if max, err := strconv.ParseInt(os.Getenv("GOTCHA_MAX_TRACERS"), 10, 64); err == nil {
		maxTracers = max
	}
	ls := golocal.LStore(maxTracers)
	// patch malloc with permanent decorator
	gomonkey.PermanentDecorate(mallocgc, func(size uintptr, tp *tp, needzero bool) unsafe.Pointer {
		// unfortunately we can't use local store direct calls here
		// as it causes `unknown caller pc` stack fatal error.
		id := ls.RLock()
		gctxPtr, ok := ls.Store[id]
		ls.RUnlock()
		if ok {
			// trace allocations for caller tracer goroutine.
			bytes := int64(size)
			objs := int64(1)
			if tp != nil {
				bytes = int64(tp.size)
				objs = int64(size) / bytes
			}
			(*gotchactx)(unsafe.Pointer(gctxPtr)).Add(bytes, objs, 1)
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
