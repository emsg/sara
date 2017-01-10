package saradb

import (
	"testing"
	"time"

	"github.com/alecthomas/log4go"
)

var (
	db  *SaraDatabase
	err error
)

func init() {
	if db, err = NewClusterDatabase("localhost:6379", 20); err != nil {
		log4go.Debug("1⃣️ ❌❌❌❌  %s", err)
		if db, err = NewDatabase("localhost:6379", 20); err != nil {
			log4go.Debug("2⃣️  ❌❌❌❌  %s", err)
		}
	}
	db.showTestLog()
}

func TestPut(t *testing.T) {
	if err != nil {
		t.Error(err)
	} else {
		for i := 0; i < 100; i++ {
			go db.Put([]byte("foo"), []byte("11111111111"))
			go db.Put([]byte("foo"), []byte("22222222222"))
			go db.Put([]byte("nono"), []byte("33333333333"))
		}
	}
}
func TestGet(t *testing.T) {
	if err != nil {
		t.Error(err)
	} else {
		foo, e0 := db.Get([]byte("foo"))
		hello, e1 := db.Get([]byte("hello"))
		nono, e2 := db.Get([]byte("nono"))
		t.Logf("foo=%s , e=%v", foo, e0)
		t.Logf("hello=%s , e=%v", hello, e1)
		t.Logf("nono=%s , e=%v", nono, e2)
	}
}

func TestPutExWithIdxAndGetByIdx(t *testing.T) {
	idx := []byte("idx")
	db.PutExWithIdx(idx, []byte("key1"), []byte("val1"), 2)
	db.PutExWithIdx(idx, []byte("key2"), []byte("val2"), 2)
	db.PutExWithIdx(idx, []byte("key3"), []byte("val3"), 2)
	time.Sleep(time.Second * 3)
	vl, err := db.GetByIdx(idx)
	t.Log(err)
	t.Logf("%s", vl)
}

func TestClose(t *testing.T) {
	db.Close()
}
