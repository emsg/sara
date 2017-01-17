package config

import (
	"testing"
)

func init() {
	SetString("foo", "bar")
	SetString("hello", "world")
}

func TestConfig(t *testing.T) {
	t.Log("==========>", GetString("foo", "foo"))
	t.Log("==========>", GetString("hello", "hello"))
	t.Log("==========>", GetString("fuck", "fuck"))

}
