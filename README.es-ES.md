# Gotcha 🎯

[![lint](https://github.com/1pkg/gotcha/workflows/lint/badge.svg)](https://github.com/1pkg/gotcha/actions?query=workflow%3Alint+branch%3Amaster+)
[![test](https://github.com/1pkg/gotcha/workflows/test/badge.svg)](https://github.com/1pkg/gotcha/actions?query=workflow%3Atest+branch%3Amaster+)
[![report](https://goreportcard.com/badge/github.com/1pkg/gotcha?nocache=1)](https://goreportcard.com/report/github.com/1pkg/gotcha?nocache=1)
[![version](https://img.shields.io/github/go-mod/go-version/1pkg/gotcha)](https://github.com/1pkg/gotcha/blob/master/go.mod)
[![license](https://img.shields.io/github/license/1pkg/gotcha)](LICENSE)
[![godoc](https://img.shields.io/badge/godoc-godoc-green)](https://pkg.go.dev/github.com/1pkg/gotcha?tab=doc)

`go get -u github.com/1pkg/gotcha`

[blog post article](https://1pkg.github.io/posts/lets_trace_goroutine_allocated_memory/)

## Introducción

Gotcha parchea el runtime de Go de manera transparente para proporcionar una forma conveniente de rastrear la cantidad de bytes asignados en el heap, objetos y llamadas por goroutine.

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
		b, o, c := ctx.Used() // bytes objetos llamadas
		fmt.Println("initial allocation", b, o, c) // imprimirá "initial allocation 824 101 2"
		gotcha.Trace(ctx, func(ctx gotcha.Context) {
			v = make([]int, 5000)
			b, o, c := ctx.Used() // bytes objetos llamadas
			fmt.Println("derived allocation", b, o, c) // imprimirá "derived allocation 40024 5001 2"
		})
		select {
		case <-ctx.Done():
			b, o, c := ctx.Used() // bytes objetos llamadas
			fmt.Println("total allocations", b, o, c) // imprimirá "total allocations 41840 5116 15"
		default:
			panic("unreachable")
		}
	}, gotcha.ContextWithLimitBytes(gotcha.KiB)) // establece el límite de asignación del contexto a un kilobit
	// ten en cuenta que los prints anteriores podrían variar ligeramente en tu máquina
	fmt.Println(len(v)) // 5000
}

```

## Internos

Gotcha expone la función `Track` que rastrea las asignaciones de memoria para una función `Tracer` proporcionada. Todas las asignaciones rastreadas se adjuntan al objeto `Context`, que es el único parámetro de esta función tracer. El contexto de Gotcha implementa completamente la interfaz `context.Context` y podría usarse para cancelar la ejecución si se superan los límites proporcionados. Gotcha admite el rastreo anidado al proporcionar el contexto de gotcha como el contexto padre para un `Tracer` derivado; entonces, los métodos del contexto de rastreo de gotcha también apuntarán al contexto padre así como al contexto derivado.

Ten en cuenta que para funcionar, gotcha utiliza los paquetes [1pkg/gomonkey](https://github.com/1pkg/gomonkey) basado en [bou.ke/monkey](https://github.com/bouk/monkey) y [1pkg/golocal](https://github.com/1pkg/golocal) basado en [modern-go/gls](https://github.com/modern-go/gls) para parchear el punto de entrada del asignador `mallocgc` del runtime y rastrear los límites de contexto por goroutine. Esto hace que gotcha herede la misma lista de restricciones que [tienen](https://github.com/bouk/monkey#notes) `modern-go/gls` y `bou.ke/monkey`.

Es importante saber que gotcha no intenta medir el uso de memoria momentáneo, lo cual implicaría el rastreo del GC (Garbage Collector); hacer un seguimiento del GC es una tarea bastante grande por sí misma y está fuera del alcance de gotcha. En su lugar, gotcha rastrea toda la memoria asignada de manera monótonamente creciente, donde solo se tienen en cuenta las asignaciones y se descartan todas las desasignaciones.

Nota: a pesar de que `gotcha` pueda funcionar en tu máquina, es extremadamente inseguro. Utiliza suposiciones de código sobre `mallocgc`, depende de convenciones de llamada que podrían cambiar, utiliza código máquina específico de la plataforma directamente, etc. Además, no se espera que `gotcha` funcione en nada que no sea `amd64` con el runtime de Go más reciente. Por lo tanto, desaconsejo encarecidamente a cualquier persona que utilice gotcha en cualquier código de producción o incluso en cualquier código en general, ya que es extremadamente inseguro, no es fiable y no contará con soporte en el futuro previsible. No obstante, uno podría imaginar casos de uso razonables para esta librería conceptual en benchmarks o pruebas de Go. Finalmente, cabe decir que la intención principal de desarrollar gotcha fue el aprendizaje y que gotcha no tiene una cobertura de pruebas exhaustiva ni soporte, y no está listo para ningún caso de uso serio de todos modos.

## Licencia

Gotcha está licenciado bajo la Licencia MIT.
Consulta [LICENSE](LICENSE) para ver el texto completo de la licencia.
