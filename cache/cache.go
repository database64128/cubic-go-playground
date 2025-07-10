package cache

import (
	"iter"
	"math"
	"sync"
	"time"
)

// Entry represents a key-value pair in the cache.
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

type expirationNode[K comparable, V any] struct {
	prev      *expirationNode[K, V]
	next      *expirationNode[K, V]
	expiresAt time.Time
	Entry[K, V]
}

// ExpirationCache is an in-memory cache with a fixed capacity where each entry has its own fixed
// expiration time. Expired entries are pruned on each set operation. When the cache reaches its
// capacity, insertions will evict the entry with the earliest expiration time (the head of the list).
type ExpirationCache[K comparable, V any] struct {
	mu        sync.RWMutex
	nodeByKey map[K]*expirationNode[K, V]
	capacity  int

	// head is the node with the earliest expiration time.
	head *expirationNode[K, V]
	// tail is the node with the latest expiration time.
	tail *expirationNode[K, V]
}

// NewExpirationCache returns a new expiration cache with the given capacity.
// If capacity is not positive, the cache will be effectively unbounded.
func NewExpirationCache[K comparable, V any](capacity int) *ExpirationCache[K, V] {
	if capacity <= 0 {
		capacity = math.MaxInt
	}
	return &ExpirationCache[K, V]{
		nodeByKey: make(map[K]*expirationNode[K, V]),
		capacity:  capacity,
	}
}

// Len returns the number of entries in the cache.
func (c *ExpirationCache[K, V]) Len() int {
	c.mu.RLock()
	length := len(c.nodeByKey)
	c.mu.RUnlock()
	return length
}

// Capacity returns the maximum number of entries the cache can hold.
func (c *ExpirationCache[K, V]) Capacity() int {
	return c.capacity
}

// Contains returns whether the cache contains the given key.
func (c *ExpirationCache[K, V]) Contains(key K) bool {
	c.mu.RLock()
	_, ok := c.nodeByKey[key]
	c.mu.RUnlock()
	return ok
}

// TryContains is like Contains, but it immediately returns false if the cache is contended.
func (c *ExpirationCache[K, V]) TryContains(key K) bool {
	if c.mu.TryRLock() {
		_, ok := c.nodeByKey[key]
		c.mu.RUnlock()
		return ok
	}
	return false
}

// Get returns the value associated with the given key, if it exists and is not expired.
func (c *ExpirationCache[K, V]) Get(key K, now time.Time) (value V, ok bool) {
	c.mu.RLock()
	node, ok := c.nodeByKey[key]
	c.mu.RUnlock()
	if !ok || !node.expiresAt.After(now) {
		return value, false
	}
	return node.Value, true
}

// GetEntry returns the entry associated with the given key, if it exists and is not expired.
func (c *ExpirationCache[K, V]) GetEntry(key K, now time.Time) (entry *Entry[K, V], ok bool) {
	c.mu.RLock()
	node, ok := c.nodeByKey[key]
	c.mu.RUnlock()
	if !ok || !node.expiresAt.After(now) {
		return nil, false
	}
	return &node.Entry, true
}

// SetFromHead inserts or updates the entry for the given key with the specified value and expiration time.
//
// As opposed to SetFromTail, SetFromHead is optimized for cases where the new expiration time is expected to be among the earliest in the cache.
// It searches for the correct insertion point starting from the head of the list (the earliest expiration).
// If the key already exists, its node is updated and repositioned; otherwise, the oldest entry is evicted if at capacity.
// All expired entries are pruned before insertion.
//
// Parameters:
//   - key:       The key to insert or update.
//   - value:     The value to associate with the key.
//   - now:       The current time, used to prune expired entries.
//   - expiresAt: The expiration time for the entry.
func (c *ExpirationCache[K, V]) SetFromHead(key K, value V, now, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node := c.getOrCreateNode(key, value, now, expiresAt)
	mark := c.head
	if mark == nil {
		c.head = node
		c.tail = node
		return
	}

	for {
		if expiresAt.Before(mark.expiresAt) {
			c.insertBefore(node, mark)
			return
		}

		mark = mark.next
		if mark == nil {
			c.insertAfter(node, c.tail)
			return
		}
	}
}

// SetFromTail inserts or updates the entry for the given key with the specified value and expiration time.
//
// As opposed to SetFromHead, SetFromTail is optimized for cases where the new expiration time is expected to be among the latest in the cache.
// It searches for the correct insertion point starting from the tail of the list (the latest expiration).
// If the key already exists, its node is updated and repositioned; otherwise, the oldest entry is evicted if at capacity.
// All expired entries are pruned before insertion.
//
// Parameters:
//   - key:       The key to insert or update.
//   - value:     The value to associate with the key.
//   - now:       The current time, used to prune expired entries.
//   - expiresAt: The expiration time for the entry.
func (c *ExpirationCache[K, V]) SetFromTail(key K, value V, now, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node := c.getOrCreateNode(key, value, now, expiresAt)
	mark := c.tail
	if mark == nil {
		c.head = node
		c.tail = node
		return
	}

	for {
		if expiresAt.After(mark.expiresAt) {
			c.insertAfter(node, mark)
			return
		}

		mark = mark.prev
		if mark == nil {
			c.insertBefore(node, c.head)
			return
		}
	}
}

// Remove deletes the value associated with the given key and returns whether the key was found.
func (c *ExpirationCache[K, V]) Remove(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.nodeByKey[key]
	if !ok {
		return false
	}
	c.remove(node)
	return true
}

// Clear removes all entries from the cache.
func (c *ExpirationCache[K, V]) Clear() {
	c.mu.Lock()
	clear(c.nodeByKey)
	c.head = nil
	c.tail = nil
	c.mu.Unlock()
}

// All returns an iterator over all entries in the cache,
// starting from the earliest expiration (head) to the latest expiration (tail).
//
// An ongoing iterator blocks concurrent writes until it completes.
func (c *ExpirationCache[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		c.mu.RLock()
		defer c.mu.RUnlock()
		for node := c.head; node != nil; node = node.next {
			if !yield(node.Key, node.Value) {
				break
			}
		}
	}
}

// Backward returns an iterator over all entries in the cache,
// starting from the latest expiration (tail) to the earliest expiration (head).
//
// An ongoing iterator blocks concurrent writes until it completes.
func (c *ExpirationCache[K, V]) Backward() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		c.mu.RLock()
		defer c.mu.RUnlock()
		for node := c.tail; node != nil; node = node.prev {
			if !yield(node.Key, node.Value) {
				break
			}
		}
	}
}

// getOrCreateNode retrieves the node for the given key if it exists, otherwise it creates a new one.
func (c *ExpirationCache[K, V]) getOrCreateNode(key K, value V, now, expiresAt time.Time) *expirationNode[K, V] {
	c.pruneExpired(now)

	node, ok := c.nodeByKey[key]
	if !ok {
		c.pruneOldest()
		node = &expirationNode[K, V]{
			expiresAt: expiresAt,
			Entry:     Entry[K, V]{Key: key, Value: value},
		}
		c.nodeByKey[key] = node
	} else {
		c.detach(node)
		node.expiresAt = expiresAt
		node.Value = value
	}

	return node
}

// pruneExpired removes all expired entries from the cache.
func (c *ExpirationCache[K, V]) pruneExpired(now time.Time) {
	for node := c.head; node != nil; node = node.next {
		if node.expiresAt.After(now) {
			break
		}
		c.remove(node)
	}
}

// pruneOldest removes the oldest entry if the cache is at capacity.
func (c *ExpirationCache[K, V]) pruneOldest() {
	if len(c.nodeByKey) < c.capacity {
		return
	}
	c.remove(c.head)
}

// insertBefore inserts the new node immediately before mark.
func (c *ExpirationCache[K, V]) insertBefore(node, mark *expirationNode[K, V]) {
	node.prev = mark.prev
	node.next = mark
	if mark.prev != nil {
		mark.prev.next = node
	} else {
		c.head = node
	}
	mark.prev = node
}

// insertAfter inserts the new node immediately after mark.
func (c *ExpirationCache[K, V]) insertAfter(node, mark *expirationNode[K, V]) {
	node.prev = mark
	node.next = mark.next
	if mark.next != nil {
		mark.next.prev = node
	} else {
		c.tail = node
	}
	mark.next = node
}

// remove deletes the given node from the cache.
func (c *ExpirationCache[K, V]) remove(node *expirationNode[K, V]) {
	delete(c.nodeByKey, node.Key)
	c.detach(node)
}

// detach detaches the given node from the list.
func (c *ExpirationCache[K, V]) detach(node *expirationNode[K, V]) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		c.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		c.tail = node.prev
	}
}
