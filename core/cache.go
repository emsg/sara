package core

import (
	"errors"
	"sync"
)

var ofl_cache *offline_message_cache

// TODO 需要用 lru 重构
type offline_message_cache struct {
	sync.RWMutex
	cache map[string][]string
}

func (self *offline_message_cache) del(key string) {
	self.Lock()
	defer self.Unlock()
	delete(self.cache, key)
}

func (self *offline_message_cache) get(key string) ([]string, error) {
	self.RLock()
	defer self.RUnlock()
	if v, ok := self.cache[key]; ok {
		return v, nil
	} else {
		return nil, errors.New("key notfound")
	}
}

func (self *offline_message_cache) put(key string, val []string) {
	self.Lock()
	defer self.Unlock()
	self.cache[key] = val
}

func init() {
	ofl_cache = &offline_message_cache{
		cache: make(map[string][]string),
	}
}
