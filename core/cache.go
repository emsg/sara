package core

import (
	"errors"
	"github.com/alecthomas/log4go"
)

var ofl_cache *offline_message_cache

// TODO 需要用 lru 重构
type offline_message_cache struct {
	cache map[string][]string
}

func (self *offline_message_cache) del(key string) {
	delete(self.cache, key)
}

func (self *offline_message_cache) get(key string) ([]string, error) {
	if v, ok := self.cache[key]; ok {
		return v, nil
	} else {
		return nil, errors.New("key notfound")
	}
}

func (self *offline_message_cache) put(key string, val []string) {
	self.cache[key] = val
}

func init() {
	log4go.Debug("📦  ofl_cache started...")
	ofl_cache = &offline_message_cache{cache: make(map[string][]string)}
}
