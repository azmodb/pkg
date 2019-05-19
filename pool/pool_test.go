package pool

import (
	"sync"
	"testing"
)

type testValue1 struct{} // testValue1 implements pool.Resetter
type testValue2 struct{}
type testValue3 struct{}

func (v *testValue1) Reset() { *v = testValue1{} }

func TestMap(t *testing.T) {
	pool1 := &sync.Pool{New: func() interface{} { return &testValue1{} }}
	pool2 := &sync.Pool{New: func() interface{} { return &testValue2{} }}
	pool3 := &sync.Pool{New: func() interface{} { return &testValue3{} }}

	m := make(Map)
	m.Register(1, pool1)
	m.Register(2, pool2)
	m.Register(3, pool3)

	if len(m) != 3 {
		t.Fatalf("map: expected length 3, have %d", len(m))
	}

	testMapGetPut(t, m)
}

func testMapGetPut(t *testing.T, m Map) {
	t.Helper()

	v1, ok := m.Get(1)
	if !ok {
		t.Fatalf("map: testValue1 pool not found")
	}
	if _, ok := v1.(*testValue1); !ok {
		t.Fatalf("map: expected testValue1, got %T", v1)
	}

	v2, ok := m.Get(2)
	if !ok {
		t.Fatalf("map: testValue2 pool not found")
	}
	if _, ok := v2.(*testValue2); !ok {
		t.Fatalf("map: expected testValue1, got %T", v2)
	}

	if ok := m.Put(1, v1); !ok {
		t.Fatalf("map: unexpected put failure")
	}
	if ok := m.Put(2, v2); !ok {
		t.Fatalf("map: unexpected put failure")
	}

	if _, ok := m.Get(42); ok {
		t.Fatalf("map: expected failure, got %v", ok)
	}
	if ok := m.Put(42, nil); ok {
		t.Fatalf("map: expected failure, got %v", ok)
	}
}
