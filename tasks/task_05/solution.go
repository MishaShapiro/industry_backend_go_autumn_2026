package main

type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	c := new(Cache[K, V])
	c.capacity = capacity
	c.items = map[K]V{}
	return c
}

func (c *Cache[K, V]) Get(k K) (v V, ok bool) {
	v, ok = c.items[k]
	return v, ok
}

func (c *Cache[K, V]) Set(k K, v V) bool {
	_, found := c.items[k]
	full := len(c.items) >= c.capacity
	if full && !found {
		return false
	}
	c.items[k] = v
	return true
}
