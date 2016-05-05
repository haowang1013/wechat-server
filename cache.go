package main

import (
	"sync"
)

type kvCache struct {
	data map[string]interface{}
	m    sync.Mutex
}

func (k *kvCache) get(key string) (interface{}, bool) {
	k.m.Lock()
	defer k.m.Unlock()
	v, ok := k.data[key]
	return v, ok
}

func (k *kvCache) set(key string, value interface{}) {
	k.m.Lock()
	defer k.m.Unlock()
	k.data[key] = value
}

func (k *kvCache) exists(key string) bool {
	k.m.Lock()
	defer k.m.Unlock()
	_, ok := k.data[key]
	return ok
}

func (k *kvCache) init() {
	if k.data == nil {
		k.data = make(map[string]interface{})
	}
}

func newCache() *kvCache {
	k := new(kvCache)
	k.init()
	return k
}
