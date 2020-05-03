package lru

import "container/list"

// Cache is a LRU cache. Not safe for concurrent access.
type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) // Callback when entry evicted
}

type entry struct {
	key   string
	value Value
}

// Value uses Len to count bytes in usage
type Value interface {
	Len() int
}

// New constrcuts a new LRU cache
func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get retrieves the entry if exists.
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// Cache hit. Update the cache
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}

	return
}

// RemoveOldest evictes the least recently used cache entry
func (c *Cache) RemoveOldest() {
	if ele := c.ll.Back(); ele != nil {
		// Remove entry from internal data structure
		val := c.ll.Remove(ele)
		kv := val.(*entry)
		delete(c.cache, kv.key)

		// Update size info
		c.nbytes -= int64(len(kv.key) + kv.value.Len())

		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add adds a value to the cache
func (c *Cache) Add(key string, value Value) {
	// Entry exists, update the existing entry
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len() - kv.value.Len())
		kv.value = value
	} else {
		newEle := c.ll.PushFront(&entry{key, value})
		c.cache[key] = newEle
		c.nbytes += int64(len(key) + value.Len())
	}

	// exceed Cache max size. evict till all entryies fit
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len returns the number of entries in the cache
func (c *Cache) Len() int {
	return c.ll.Len()
}
