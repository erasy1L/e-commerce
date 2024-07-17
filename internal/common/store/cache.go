package store

import (
	"encoding/json"
	"sync"
)

type Cache[T any] interface {
	Get(key string) (T, bool)
	Set(key string, value T)
}

type InMemoryIdempotencyCache[T any] struct {
	cache sync.Map
}

func NewInMemoryIdempotencyCache[T any]() *InMemoryIdempotencyCache[T] {
	return &InMemoryIdempotencyCache[T]{}
}

func (c *InMemoryIdempotencyCache[T]) Get(key string) (val T, has bool) {
	v, has := c.cache.Load(key)
	if !has {
		return
	}

	err := json.Unmarshal(v.([]byte), &val)
	if err != nil {
		return
	}

	return val, has
}

func (c *InMemoryIdempotencyCache[T]) Set(key string, value T) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	c.cache.Store(key, data)
}
