package wire

import (
	"math"
	"testing"
)

type testStruct struct {
	Uint64 uint64
	Uint32 uint32
	Uint16 uint16
	Uint8  uint8
	String string
}

func TestSizeOf(t *testing.T) {
	t.Parallel()

	for i, testcase := range []struct {
		src  interface{}
		want int
	}{
		{[]testStruct{
			testStruct{math.MaxUint64, math.MaxUint32, math.MaxUint16, math.MaxUint8, "hello world"},
			testStruct{math.MaxUint64, math.MaxUint32, math.MaxUint16, math.MaxUint8, "hello world"},
			testStruct{math.MaxUint64, math.MaxUint32, math.MaxUint16, math.MaxUint8, "hello world"},
		}, 86},
		{[]testStruct{}, 2},

		{testStruct{math.MaxUint64, math.MaxUint32, math.MaxUint16, math.MaxUint8, "hello world"}, 28},
		{testStruct{}, 17},

		{[]string{"a", "b", "c", "d"}, 14},
		{[]string{"", "", "", ""}, 10},

		{[]byte("abcd"), 8},
		{[]byte(""), 4},

		{"abcd", 6},
		{"", 2},

		{uint64(math.MaxUint64), 8},
		{uint32(math.MaxUint32), 4},
		{uint16(math.MaxUint16), 2},
		{uint8(math.MaxUint8), 1},
		{uint64(0), 8},
		{uint32(0), 4},
		{uint16(0), 2},
		{uint8(0), 1},
	} {
		size := SizeOf(testcase.src)
		if size != testcase.want {
			t.Errorf("sizeof (%.4d): expected size %d, got %d", i, testcase.want, size)
		}
	}
}
