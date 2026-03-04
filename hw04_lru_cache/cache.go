package hw04lrucache

import (
	"errors"
	"sync"
)

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mu       *sync.Mutex
}

func NewCache(capacity int) Cache {
	var mu sync.Mutex
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
		mu:       &mu,
	}
}

// Метод удаления из списка и из кеша.
func (lru *lruCache) delitem(item *ListItem) error {
	for k, v := range lru.items {
		if item == v {
			lru.queue.Remove(item)
			delete(lru.items, k)
			return nil
		}
	}
	return errors.New("element not exist in items")
}

func (lru *lruCache) Set(key Key, value interface{}) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	item, ok := lru.items[key]
	if ok {
		item.Value = value
		lru.queue.MoveToFront(item)
		lru.items[key] = item
		return true
	}
	if lru.capacity == lru.queue.Len() {
		last := lru.queue.Back()
		err := lru.delitem(last)
		if err != nil {
			panic(err.Error())
		}
		item = lru.queue.PushFront(value)
	} else {
		item = lru.queue.PushFront(value)
	}
	lru.items[key] = item

	return false
}

func (lru *lruCache) Get(key Key) (interface{}, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	item, ok := lru.items[key]
	if ok {
		lru.queue.MoveToFront(item)
		return item.Value, true
	}
	return nil, false
}

func (lru *lruCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.items = make(map[Key]*ListItem, lru.capacity)
	lru.queue = NewList()
}
