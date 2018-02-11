package cache

import (
	"time"
	"sync"
	"errors"
)

// Config add some setting for cache
type Config struct {
	DefaultExpire time.Duration
	ClearInterval time.Duration
}

// Cache is main type
type Cache struct {
	sync.RWMutex
	expire time.Duration
	interval time.Duration
	items map[string]*item
}

type item struct {
	data interface{}
	expire time.Time
}

const (
	// Forever means what item will newer be expired
	Forever time.Duration = 0
	// Default expired 
	Default time.Duration = -1
)

// NewCache is create and return new cache
func NewCache(c Config) *Cache {
	cache := &Cache{}
	cache.items = map[string]*item{}
	cache.interval = c.ClearInterval
	cache.expire = c.DefaultExpire

	if cache.interval != -1 {
		go cache.GCtick()
	}

	return cache
}
// Add new key-value pair 
func (c *Cache) Add(k string, v interface{}, d time.Duration) (bool, error) {
	if ok := c.Has(k); ok {
		return false, errors.New("Key is already exist")
	}
	c.Lock()
	c.items[k] = c.set(k, v, d)
	c.Unlock()
	return true, nil
}
// Set item 
func (c *Cache) Set(k string, v interface{}, d time.Duration) (bool, error) {
	c.Lock()
	c.items[k] = c.set(k, v, d)
	c.Unlock()
	return true, nil
}
// Update value by key
func(c *Cache) Update(k string, v interface{}, d time.Duration) (bool, error) {
	if ok := c.Has(k); !ok {
		return false, errors.New("Key not found")
	}
	c.Lock()
	c.items[k].data = v
	if d != -1 {
		c.items[k].expire = time.Now().Add(d)
	}
	c.Unlock()
	return true, nil
}

func (c *Cache) set(k string, v interface{}, d time.Duration) *item {
	item := &item{}
	item.data = v

	if d == Default {
		d = c.expire
	}

	if d != Forever {
		item.expire = time.Now().Add(d)
	}

	return item
}
// Has check is key exsist
func (c *Cache) Has(k string) bool {
	c.RLock()
	ok := c.has(k)
	c.RUnlock()
	return ok
}

func (c *Cache) has(k string) bool {
	_, ok := c.items[k]
	return ok
}
// Get values by keys
func (c *Cache) Get(k []string) ([]interface{}, bool) {
	if len(k) < 1 {
		return nil, false
	}
	c.RLock()
	item, ok := c.get(k)
	c.RUnlock()
	return item, ok
}

func (c *Cache) get(k []string) ([]interface{}, bool) {
	
	var i []interface{}

	for _, key := range k{
		item, ok := c.items[key]
		if ok && !item.expired() {
			i = append(i, item.data)
		}
	}
	if len(i) > 0 {
		return i, true
	}

	return nil, false
}

// Del delete an item from cache by key
func (c *Cache) Del(k string) {
	c.Lock()
	c.del(k)
	c.Unlock()
}

func (c *Cache) del(k string) {
	delete(c.items, k)
}

func (i *item) expired() bool {
	if !i.expire.IsZero() {
		return time.Now().After(i.expire)
	}
	return false
}

func (c *Cache) Clean() {
	c.Lock()
	c.clean()
	c.Unlock()
}

func (c *Cache) clean() {
	for k, v := range c.items {
		if v.expired() {
			c.del(k)
		}
	}
}
func (c *Cache) GCtick() {
	for _ = range time.Tick(c.interval) {
		c.Clean()
	}
}