package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v4"
	"sync"
	"time"
)

type kvCache interface {
	get(key string) (string, bool)
	set(key, value string) error
	exists(key string) bool
}

type factory func() interface{}

func getJson(k kvCache, key string, f factory) (interface{}, bool) {
	s, ok := k.get(key)
	if !ok {
		return nil, false
	}

	log.Debugf("%s: %s", key, s)

	if s == "null" {
		return nil, true
	}

	value := f()
	err := json.Unmarshal([]byte(s), value)
	logError(err)
	return value, err == nil
}

func setJson(k kvCache, key string, value interface{}) error {
	buff, err := json.Marshal(value)
	logError(err)
	if err != nil {
		return err
	}
	return k.set(key, string(buff))
}

func logError(err error) {
	if err != nil {
		log.Error(err)
	}
}

/**
* memory cache
 */
type memCache struct {
	data map[string]string
	m    sync.Mutex
}

func (k *memCache) init() {
	if k.data == nil {
		k.data = make(map[string]string)
	}
}

func (k *memCache) get(key string) (string, bool) {
	k.m.Lock()
	defer k.m.Unlock()
	v, ok := k.data[key]
	return v, ok
}

func (k *memCache) set(key, value string) error {
	k.m.Lock()
	defer k.m.Unlock()
	k.data[key] = value
	return nil
}

func (k *memCache) exists(key string) bool {
	k.m.Lock()
	defer k.m.Unlock()
	_, ok := k.data[key]
	return ok
}

func newMemCache() kvCache {
	k := new(memCache)
	k.init()
	return k
}

/**
* redis cache
 */
type redisCache struct {
	client        *redis.Client
	keyPrefix     string
	valueLifeTime time.Duration
}

func (r *redisCache) init(address, keyPrefix string, valueLifeTime time.Duration) {
	r.client = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := r.client.Ping().Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect to redis server '%s': %s", address, err))
	}
	r.keyPrefix = keyPrefix
	r.valueLifeTime = valueLifeTime
}

func (r *redisCache) get(key string) (string, bool) {
	modKey := r.getKey(key)
	s, err := r.client.Get(modKey).Result()
	if err == redis.Nil {
		return "", false
	}

	logError(err)
	if err == nil {
		return s, true
	} else {
		return "", false
	}
}

func (r *redisCache) set(key, value string) error {
	modKey := r.getKey(key)
	c := r.client.Set(modKey, value, r.valueLifeTime)
	err := c.Err()
	logError(err)
	return err
}

func (r *redisCache) getKey(key string) string {
	return r.keyPrefix + "." + key
}

func (r *redisCache) exists(key string) bool {
	modKey := r.getKey(key)
	return r.client.Exists(modKey).Val()
}

func newRedisCache(address, keyPrefix string, valueLifeTime time.Duration) kvCache {
	r := new(redisCache)
	r.init(address, keyPrefix, valueLifeTime)
	return r
}
