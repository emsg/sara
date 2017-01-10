package saradb

import "testing"

func TestSub(t *testing.T) {
	t.Log("✈️ ✈️ ✈️ ✈️  ------>")
	if db, err := NewDatabase("127.0.0.1:6379", 20); err == nil {
		channel := db.GenDataChannel("fuck")
		sc := make(chan string)
		channel.Subscribe(func(message string) { sc <- message })
		channel.Publish(channel.GetChannel(), "hello world.")
		m := <-sc
		t.Log("sub_msg =", m)
	} else {
		t.Error(err)
	}
}
