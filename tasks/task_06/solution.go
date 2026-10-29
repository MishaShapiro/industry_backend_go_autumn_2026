package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}

type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	c := new(LRUCache[K, V])
	c.capacity = capacity
	c.items = map[K]*list.Element{}
	return c
}

func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	node, ok := c.items[key]
	if !ok {
		return value, false
	}
	c.ll.MoveToFront(node)
	e := node.Value.(*entry[K, V])
	return e.value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}

	if node, ok := c.items[key]; ok {
		e := node.Value.(*entry[K, V])
		e.value = value
		return
	}

	if len(c.items) == c.capacity {
		c.removeOldest()
	}
	c.items[key] = c.ll.PushFront(&entry[K, V]{key: key, value: value})
}

func (c *LRUCache[K, V]) removeOldest() {
	node := c.ll.Back()
	if node == nil {
		return
	}
	e := c.ll.Remove(node).(*entry[K, V])
	delete(c.items, e.key)
}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
