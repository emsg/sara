package external

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestParams(t *testing.T) {
	p := NewParams("emsg_auth", "auth", "uid", "foo", "pwd", "bar")
	t.Log(p.ToJson())
}

func TestJson(t *testing.T) {
	r1 := gjson.Parse("{foo:1,bar:2")
	t.Log(r1.Exists(), r1.Type)
	t.Log(gjson.Get("{foo:1", "foo"))
}
