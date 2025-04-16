// A [sync.Map] with generics.
package syncmap

import (
	"sync"
)

type SyncMap[K comparable, V any] struct {
	sync.Map
}

func New[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		Map: sync.Map{},
	}
}

func (c *SyncMap[K, V]) Load(key K) (V, bool) {
	v, ok := c.Map.Load(key)
	return v.(V), ok
}

func (c *SyncMap[K, V]) Store(key K, value V) {
	c.Map.Store(key, value)
}

func (c *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	c.Map.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}
