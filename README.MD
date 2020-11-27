# Gotcha 🎯

[![lint](https://github.com/1pkg/gotcha/workflows/lint/badge.svg)](https://github.com/1pkg/gotcha/actions?query=workflow%3Alint+branch%3Amaster+)
[![test](https://github.com/1pkg/gotcha/workflows/test/badge.svg)](https://github.com/1pkg/gotcha/actions?query=workflow%3Atest+branch%3Amaster+)
[![report](https://goreportcard.com/badge/github.com/1pkg/gotcha)](https://goreportcard.com/report/github.com/1pkg/gotcha)
[![version](https://img.shields.io/github/go-mod/go-version/1pkg/gotcha)](https://github.com/1pkg/gotcha/blob/master/go.mod)
[![license](https://img.shields.io/github/license/1pkg/gotcha)](LICENSE)
[![godoc](https://img.shields.io/badge/godoc-godoc-green)](https://pkg.go.dev/github.com/1pkg/gotcha?tab=doc)

`go get -u github.com/1pkg/gotcha`

## Introduction

Gotcha seamlessly patches go runtime to provides a way to track amount of allocated bytes, objects, calls per goroutine.

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

Gotcha exposes function `Track` that tracks memory allocations for provided `Tracer` function. All traced allocations are attacehd to the single parameter of this tracer function `Context` object. Gotcha context fully implements `context.Context` interface and could be used to cancel execution if provided limits were exceeded. Gotcha supports nested treacking by providing gotcha context as parent context for derived `Tracer` function then gotcha context `Tracker` methods `Add`, `Remains` and `Exceeded` will also target parent context as well which is useful if nested tracking is required.
Note that in order to work gotcha uses [bou.ke/monkey](https://github.com/bouk/monkey) and [modern-go/gls](https://github.com/modern-go/gls) packages to patch existing runtime allocator entrypoints and track per goroutine context limts, next entrypoints are patched:

- direct objects allocation
- arrays allocation
- slice allocation
- map allocation (solved by arrays allocation)
- chan allocation
- strings/bytes/runes allocation

This makes gotcha inherits the same list of restrictions as `modern-go/gls` and `bou.ke/monkey` [has](https://github.com/bouk/monkey#notes).

Note that some patches with monkey patch are causing loops, for e.g. `newarray`, `growslice`.
So for them implementation either slightly changed - newarray or not patched at all - `growslice`.
For `growslice` it still possible to make the same workaround as done for `newarray`, but it will require to copy and support great amount of code from runtime which is not correlating with the goal of this project, so `growslice` is skipped for now. Note that some function from `interface.conv` family are not patched neither which will cause untracked allocation for code like `vt, ok := var.(type)`.
Note that `runtime.gobytes` is not patched as well as it seems it's only used by go compiler itself.
Note that only functions from `mallocgc` family are patched, but runtime has much more allocation tricks
that won't be traced by gotcha, like direct `malloc` sys calls, etc.

## Licence

Gotcha is licensed under the MIT License.  
See [LICENSE](LICENSE) for the full license text.
