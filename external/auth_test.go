package external

import "sara/config"
import "testing"

func TestAuth(t *testing.T) {
	url := "http://127.0.0.1:8000/request/"
	config.SetString("callback", url)
	a := Auth("foo", "bar")
	t.Log(a)
}
