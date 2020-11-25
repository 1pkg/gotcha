package main

import "fmt"

type Context struct {
	Bytes, Objects, Calls int64
}

func (ctx Context) String() string {
	return fmt.Sprintf(
		"on this context: %d objects has been allocated with total size of %d bytes within %d calls",
		ctx.Objects,
		ctx.Bytes,
		ctx.Calls,
	)
}

func (ctx *Context) add(bytes, objects int64) {
	if objects > 0 && bytes > 0 {
		ctx.Bytes += bytes * objects
		ctx.Objects += objects
		ctx.Calls++
	}
}
