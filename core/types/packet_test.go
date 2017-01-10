package types

import (
	"testing"
)

func TestJID(t *testing.T) {
	jid, _ := NewJID("foo@bar.com/helloworld")
	t.Log(jid)
	t.Log(jid.StringWithoutResource())
}

func TestDecodeAndEncode(t *testing.T) {
	jsonData := []byte(`{"envelope":{"id":"1234567890","type":0,"jid":"usera@test.com","pwd":"abc123"},"vsn":"0.0.1"}`)
	if p, e := NewPacket(jsonData); e != nil {
		t.Error(e)
	} else {
		t.Log(p)
		t.Log(p.Envelope.Id == "1234567890")
		data := p.ToJson()
		t.Logf("%s", data)
	}

}
