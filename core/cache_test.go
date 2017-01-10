package core

import "testing"

func TestOflcache(t *testing.T) {
	ofl_cache.put("1", []string{"a0", "b", "c"})
	ofl_cache.put("2", []string{"a1", "b", "c"})
	ofl_cache.put("3", []string{"a2", "b", "c"})
	t.Log(ofl_cache.cache)
	r, ok := ofl_cache.get("2")
	t.Log(ok, r)
	ofl_cache.del("2")
	r, ok = ofl_cache.get("2")
	t.Log(ok, r)
}
