package pool

import "testing"

func TestGeneratorLimit(t *testing.T) {
	p := NewGenerator(1, 1000)
	for i := 0; i < 999; i++ {
		if _, ok := p.Get(); !ok {
			t.Fatalf("generator: limit reached after %d values", i)
		}
	}

	if _, ok := p.Get(); ok {
		t.Fatalf("generator: not exhausted when it should be")
	}
}

func TestGeneratorUnique(t *testing.T) {
	got := make(map[int64]bool)
	p := NewGenerator(1, 5)

	for {
		v, ok := p.Get()
		if !ok {
			break
		}

		if _, ok := got[v]; ok {
			t.Fatalf("generator: found duplicate value %d", v)
		}
		got[v] = true
	}
}

func TestGeneratorRecycle(t *testing.T) {
	p := NewGenerator(1, 16)
	v1, _ := p.Get()
	p.Put(v1)
	v2, _ := p.Get()
	if v1 != v2 {
		t.Fatalf("generator: pool not recycled values")
	}
}
