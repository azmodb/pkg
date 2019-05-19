package pool

import "sync"

// Generator represents a numeric identifier allocator. It can be used for
// both tags and fids.
//
// A Generator is safe for use by multiple goroutines simultaneously.
type Generator struct {
	mu    sync.Mutex
	m     []int64
	cur   int64
	limit int64
}

// NewGenerator returns a new numeric identifier allocator. Start is the
// starting value and limit is the upper limit.
func NewGenerator(start int64, limit int64) *Generator {
	return &Generator{cur: start, limit: limit}
}

// Get gets a value from the pool.
func (g *Generator) Get() (int64, bool) {
	g.mu.Lock()
	if len(g.m) > 0 {
		v := g.m[len(g.m)-1]
		g.m = g.m[:len(g.m)-1]
		g.mu.Unlock()
		return v, true
	}
	if g.cur == g.limit {
		g.mu.Unlock()
		return 0, false
	}
	v := g.cur
	g.cur++
	g.mu.Unlock()
	return v, true
}

// Put returns the value to the pool.
func (g *Generator) Put(v int64) {
	g.mu.Lock()
	g.m = append(g.m, v)
	g.mu.Unlock()
}
