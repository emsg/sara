package sararpc

import (
	"fmt"
	"testing"

	"github.com/golibs/uuid"
)

func BenchmarkCallSararpc(b *testing.B) {
	l := uuid.Rand().Hex()
	c := NewRPCClient()
	p := `{"envelope":{"id":"xxxx","type":1,"from":"usera@test.com","to":"userb@test.com","ack":1},"vsn":"0.0.1","payload":{"content":"hellow world","attrs":{"lat":"a.a","lng":"b.b"}}}`
	for i := 0; i < b.N; i++ {
		c.Call("127.0.0.1:4281", l, fmt.Sprintf("s-%d", i), p)
	}
}
