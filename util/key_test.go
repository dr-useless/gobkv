package util

import (
	"bytes"
	"fmt"
	"testing"
)

func TestFastXor(t *testing.T) {
	dst := make([]byte, 16)
	a := []byte{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1}
	b := []byte{1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}
	FastXor(dst, a, b)

	// test that dst == res
	res := []byte{0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0}
	if !bytes.Equal(dst, res) {
		t.FailNow()
	}
}

func BenchmarkXorLoop(bench *testing.B) {
	dst := make([]byte, 16)
	a := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1}
	b := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}

	bench.StartTimer()
	for i := 0; i < bench.N; i++ {
		SlowXor(dst, a, b)
	}
	bench.StopTimer()

	fmt.Println(dst)
}

func BenchmarkFastXor(bench *testing.B) {
	dst := make([]byte, 16)
	a := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1}
	b := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1}

	bench.StartTimer()
	for i := 0; i < bench.N; i++ {
		FastXor(dst, a, b)
	}
	bench.StopTimer()

	fmt.Println(dst)
}
