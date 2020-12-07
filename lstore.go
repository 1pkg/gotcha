package gotcha

import (
	"sync/atomic"

	"github.com/modern-go/gls"
)

var glstore *lstore

type lstore struct {
	mp   map[int64]Context
	cap  int64
	lock int64
}

func (ls *lstore) get() Context {
	if i := atomic.LoadInt64(&ls.lock); i == 0 {
		return ls.mp[gls.GoID()]
	}
	return nil
}

func (ls *lstore) set(ctx Context) {
	atomic.StoreInt64(&ls.lock, 1)
	defer atomic.StoreInt64(&ls.lock, 0)
	if int64(len(ls.mp)) == ls.cap {
		return
	}
	ls.mp[gls.GoID()] = ctx
}

func (ls *lstore) del() {
	atomic.StoreInt64(&ls.lock, 1)
	defer atomic.StoreInt64(&ls.lock, 0)
	delete(ls.mp, gls.GoID())
}
