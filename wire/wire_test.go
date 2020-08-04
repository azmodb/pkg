package wire

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"testing"
)

func allocType(t *testing.T, src interface{}) (dst interface{}) {
	t.Helper()

	srcType := reflect.TypeOf(src)
	if srcType.Kind() == reflect.Ptr {
		dst = reflect.New(srcType.Elem()).Interface()
	} else {
		dst = reflect.New(srcType).Interface()
	}
	return dst
}

func deepEqualType(t *testing.T, src, dst interface{}) bool {
	t.Helper()

	a, b := reflect.ValueOf(src), reflect.ValueOf(dst)
	if a.Type().Kind() != reflect.Ptr {
		b = reflect.Indirect(b)
	}
	return reflect.DeepEqual(a.Interface(), b.Interface())
}

func TestMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	b := NewBuffer(nil)
	for i, testcase := range []struct {
		src interface{}
	}{
		{src: []testStruct{
			testStruct{math.MaxUint64, math.MaxUint32, math.MaxUint16, math.MaxUint8, "hello world"},
			testStruct{math.MaxUint64, math.MaxUint32, math.MaxUint16, math.MaxUint8, "hello world"},
			testStruct{math.MaxUint64, math.MaxUint32, math.MaxUint16, math.MaxUint8, "hello world"},
		}},
		// {src: []testStruct{}},

		{src: testStruct{math.MaxUint64, math.MaxUint32, math.MaxUint16, math.MaxUint8, "hello world"}},
		{src: testStruct{}},

		{src: []string{"a", "b", "c", "d"}},
		//{src: []string{}},

		{src: []byte("hello world")},
		{src: "hello world"},

		{src: uint64(math.MaxUint64)},
		{src: uint32(math.MaxUint32)},
		{src: uint16(math.MaxUint16)},
		{src: uint8(math.MaxUint8)},
		{src: uint64(42)},
		{src: uint32(42)},
		{src: uint16(42)},
		{src: uint8(42)},
	} {
		b.Reset()
		if err := b.Marshal(testcase.src); err != nil {
			t.Errorf("marshal (%.4d): %v", i, err)
		}
		size := SizeOf(testcase.src)
		if b.Len() != SizeOf(testcase.src) {
			t.Errorf("marshal (%.4d): expected marshaled size %d, got %d", i, size, b.Len())
		}

		dst := allocType(t, testcase.src)
		if err := b.Unmarshal(dst); err != nil {
			t.Errorf("unmarshal (%.4d): %v", i, err)
		}

		if !deepEqualType(t, testcase.src, dst) {
			t.Errorf("marshal/unmarshal (%.4d):\nwant %#v\ngot  %#v", i, testcase.src, dst)
		}
		if b.Len() != 0 {
			t.Errorf("marshal/unmarshal (%.4d): expected empty buffer, got %d", i, b.Len())
		}
	}
}

func TestParseError(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, desc string, n, want int) {
		t.Helper()
		if n != want {
			t.Errorf("%s: expected error code %d, got %d", desc, want, n)
		}
	}

	_, n := ConsumeBytes(nil, nil)
	check(t, "ConsumeBytes", n, errUnexpectedEOF)

	_, n = ConsumeString(nil)
	check(t, "ConsumeString", n, errUnexpectedEOF)

	_, n = ConsumeUint64(nil)
	check(t, "ConsumeUint64", n, errUnexpectedEOF)

	_, n = ConsumeUint32(nil)
	check(t, "ConsumeUint32", n, errUnexpectedEOF)

	_, n = ConsumeUint16(nil)
	check(t, "ConsumeUint16", n, errUnexpectedEOF)

	_, n = ConsumeUint8(nil)
	check(t, "ConsumeUint8", n, errUnexpectedEOF)
}

func TestWriteTo(t *testing.T) {
	str := "NOTE 3: Each bit has the value either ZERO or ONE. --- ECMA-035 spec"

	var buf bytes.Buffer
	b := NewBuffer(nil)
	b.WriteString(str)

	n, err := b.WriteTo(&buf)
	if err != nil {
		t.Fatalf("writeto: %v", err)
	}
	if int(n) != len(str) {
		t.Fatalf("writeto: expected %d written bytes, got %d", len(str), n)
	}
	if buf.String() != str {
		t.Fatalf("writeto: expected content %q, got %q", str, buf.String())
	}
	if b.Len() != 0 {
		t.Fatalf("writeto: expected empty buffer, got %d", b.Len())
	}

	if _, err = b.WriteTo(&buf); err != io.EOF {
		t.Fatalf("writeto: expected EOF error, got %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func TestRead(t *testing.T) {
	b := NewBuffer(nil)
	for i, testcase := range []struct {
		data string

		p   []byte
		err error
	}{
		{"abcd", make([]byte, 4), nil},
		{"abcd", make([]byte, 3), nil},
		{"abcd", make([]byte, 2), nil},
		{"abcd", make([]byte, 1), nil},
		{"abcd", make([]byte, 0), nil},

		{"abcd", make([]byte, 5), nil},
		{"abcd", make([]byte, 6), nil},

		{"", make([]byte, 6), io.EOF},
	} {
		b.Reset()
		b.WriteString(testcase.data)

		n, err := b.Read(testcase.p)
		if err != testcase.err {
			t.Errorf("read (%.4d): expected %v error, got %v", i, testcase.err, err)
		}
		want := min(len(testcase.data), len(testcase.p))
		if n != want {
			t.Errorf("read (%.4d): expected %d read bytes, got %d", i, want, n)
		}
		want = max(len(testcase.data)-len(testcase.p), 0)
		if b.Len() != want {
			t.Errorf("read (%.4d): expected buffer size %d, got %d", i, want, b.Len())
		}
	}
}

func TestCopy(t *testing.T) {
	str := "RULE 7: Option arguments cannot be optional. --- Sys V Interface p 343."

	var buf bytes.Buffer
	b := NewBuffer(nil)
	b.WriteString(str)

	n, err := io.Copy(&buf, b)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if int(n) != len(str) {
		t.Fatalf("copy: expected %d written bytes, got %d", len(str), n)
	}
	if b.Len() != 0 {
		t.Fatalf("copy: expected empty buffer, got %d", b.Len())
	}
	if buf.String() != str {
		t.Fatalf("copy: expected content %q, got %q", str, buf.String())
	}
}
