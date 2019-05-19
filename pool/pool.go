package pool

import "github.com/azmodb/pkg/log"

// LimitPool is a set of temporary objects that may be individually saved
// and retrieved.
//
// LimitPool's purpose is to cache up to Limit allocated but unused
// values for later reuse. That is, it makes it easy to build efficient
// and memory limited, thread-safe free lists.
type LimitPool struct {
	Factory func() interface{}
	Limit   int

	cache chan interface{}
}

// DefaultLimit is the default maximal cache size.
const DefaultLimit = 8

func (p *LimitPool) init() {
	if p.cache == nil {
		if p.Limit <= 0 {
			p.Limit = DefaultLimit
		}
		p.cache = make(chan interface{}, p.Limit)
	}
}

// Get selects an arbitrary value from the pool, removes it from the pool
// and returns it to the caller.
func (p *LimitPool) Get() (value interface{}) {
	p.init()

	select {
	case value = <-p.cache:
	default:
		if p.Factory == nil {
			log.Panicf("pool: LimitPool factory function not set")
		}
		value = p.Factory()
	}
	return value
}

// Put returns the value to the pool.
func (p *LimitPool) Put(value interface{}) {
	p.init()

	select {
	case p.cache <- value:
	default:
	}
}

// Pool represents a set of temporary objects that may be individually
// saved and retrieved.
//
// Pool must be safe for use by multiple goroutines simultaneously.
type Pool interface {
	// Get selects an arbitrary value from the Pool, removes it from
	// the Pool, and returns it to the caller.
	Get() interface{}

	// Put returns the value to the pool.
	Put(interface{})
}

// Map represents a Pool registry.
type Map map[interface{}]Pool

// Get selects an arbitrary value from the Pool, removes it from
// the Pool, and returns it to the caller.
//
// The success result indicates whether a pool was found in the pool
// map.
func (m Map) Get(key interface{}) (value interface{}, success bool) {
	pool, success := m[key]
	if !success {
		return value, false
	}
	value = pool.Get()
	return value, success
}

// Resetter resets all receiver state.
type Resetter interface {
	Reset()
}

// Put returns the value to the pool. If value implements Resettter,
// Put resets value.
//
// The success result indicates whether a pool was found in the pool
// map.
func (m Map) Put(key interface{}, value interface{}) bool {
	pool, success := m[key]
	if !success {
		return false
	}
	if r, ok := value.(Resetter); ok {
		r.Reset()
	}
	pool.Put(value)
	return true
}

// Register sets the pool for a key. Register is not safe for use by
// multiple goroutines simultaneously.
//
// Register should only be used from init().
func (m *Map) Register(key interface{}, pool Pool) {
	if _, found := (*m)[key]; found {
		log.Panicf("map: found duplicate pool identifier: <%v>", key)
	}
	(*m)[key] = pool
}
