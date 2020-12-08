# Gotcha 🎯

[![lint](https://github.com/1pkg/gotcha/workflows/lint/badge.svg)](https://github.com/1pkg/gotcha/actions?query=workflow%3Alint+branch%3Amaster+)
[![test](https://github.com/1pkg/gotcha/workflows/test/badge.svg)](https://github.com/1pkg/gotcha/actions?query=workflow%3Atest+branch%3Amaster+)
[![report](https://goreportcard.com/badge/github.com/1pkg/gotcha)](https://goreportcard.com/report/github.com/1pkg/gotcha)
[![version](https://img.shields.io/github/go-mod/go-version/1pkg/gotcha)](https://github.com/1pkg/gotcha/blob/master/go.mod)
[![license](https://img.shields.io/github/license/1pkg/gotcha)](LICENSE)
[![godoc](https://img.shields.io/badge/godoc-godoc-green)](https://pkg.go.dev/github.com/1pkg/gotcha?tab=doc)

`go get -u github.com/1pkg/gotcha`

## Introduction

Gotcha seamlessly patches go runtime to provide a convenient way to track amount of heap allocated bytes, objects, calls per goroutine.

```go
package main

import (
	"context"
	"fmt"

	"github.com/1pkg/gotcha"
)

func main() {
	var v []int
	gotcha.Trace(context.Background(), func(ctx gotcha.Context) {
		v = make([]int, 100)
		b, o, c := ctx.Used() // bytes objects calls
		fmt.Println("initial allocation", b, o, c) // will print "initial allocation 824 101 2"
		gotcha.Trace(ctx, func(ctx gotcha.Context) {
			v = make([]int, 5000)
			b, o, c := ctx.Used() // bytes objects calls
			fmt.Println("derived allocation", b, o, c) // will print "derived allocation 40024 5001 2"
		})
		select {
		case <-ctx.Done():
			b, o, c := ctx.Used() // bytes objects calls
			fmt.Println("total allocations", b, o, c) // will print "total allocations 41840 5116 15"
		default:
			panic("unreachable")
		}
	}, gotcha.ContextWithLimitBytes(gotcha.KiB)) // set context allocation limit to one kilobit
	// note that prints above might be slightly different on your machine
	fmt.Println(len(v)) // 5000
}

```

## Internals

Gotcha exposes function `Track` that tracks memory allocations for provided `Tracer` function. All traced allocations are attached to the single parameter of this tracer function `Context` object. Gotcha context fully implements `context.Context` interface and could be used to cancel execution if provided limits were exceeded. Gotcha supports nested tracing by providing gotcha context as the parent context for derived `Tracer`; then gotcha tracing context methods will also be targeting parent context as well as derived context.

Note that in order to work gotcha uses [1pkg/gomonkey](https://github.com/1pkg/gomonkey) based on [bou.ke/monkey](https://github.com/bouk/monkey) and [1pkg/golocal](https://github.com/1pkg/golocal) based on [modern-go/gls](https://github.com/modern-go/gls) packages to patch runtime `mallocgc` allocator entrypoint and trace per goroutine context limts. This makes gotcha inherits the same list of restrictions as `modern-go/gls` and `bou.ke/monkey` [has](https://github.com/bouk/monkey#notes).

It's important to know that gotcha is not trying to measure momentary memory usage which involves GC tracing into the act, keeping track on GC is rather a big task on it's own and out of scope for gotcha. Instead gotcha traces all memory allocated in monotonic increasing fashion where is only allocations are taken into consideration and all deallocations are discarded.

Note: despite that `gotcha` might work on your machine, it's super unsafe. It uses code assumption about `mallocgc`, depends on calling convention that could be changed, uses platform specific machine code direcly, etc. Alsp `gotcha` isn't expected to work on anything apart from `amd64` with latest go runtime. So I highly discourage anyone to use gotcha in any production code or maybe even any code at all as it's extremely unsafe and not reliable and won't be supported in foreseeable future. Nevertheless, one could probably imagine reasonable use cases for this concept library in go benchmarks or test. Finally it's worth to say that primary intention to develop gotcha was learning and that gotcha doesn't have comprehensive tests coverage and support and not ready for any serious use case anyway

## Licence

Gotcha is licensed under the MIT License.  
See [LICENSE](LICENSE) for the full license text.
