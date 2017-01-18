package external

import "sara/config"
import "testing"

var packetStr string = `{"envelope":{"id":"1","ack":1,"gid":"g123","from":"usera@test.com","type":2,"ct":"1401331382839","to":"a3@test.com"},"payload":{"content":"hi"}}`

func TestOfflineCallback(t *testing.T) {
	url := "http://127.0.0.1:8000/request/"
	config.SetString("callback", url)
	a := OfflineCallback(packetStr)
	t.Log(a)
}
