package gotcha

import (
	"testing"
)

func TestTraceTypes(t *testing.T) {
	var v []int64
	for i := 0; i < 100; i++ {
		v = make([]int64, 1, 100)
	}
	v[0] = 0

}
