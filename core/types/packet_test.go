package types

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/gjson"
)

func TestJID(t *testing.T) {
	jid, _ := NewJID("foo@bar.com/helloworld")
	t.Log(jid)
	t.Log(jid.StringWithoutResource())
}

var jsonData []byte = []byte(`{"envelope":{"id":"1234567890","type":0,"jid":"usera@test.com","pwd":"abc123"},"vsn":"0.0.1","payload":{"content":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`)

func TestDecodeAndEncode(t *testing.T) {
	if p, e := NewPacket(jsonData); e != nil {
		t.Error(e)
	} else {
		t.Log(p.Envelope.Id)
		t.Log(p.Envelope.Id == "1234567890")
		data := p.ToJson()
		t.Logf("%s", data)
	}

}

func BenchmarkJson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var envelope interface{}
		json.Unmarshal(jsonData, &envelope)
	}
}
func BenchmarkGson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var envelope interface{}
		e := gjson.Get(string(jsonData), "envelope")
		json.Unmarshal([]byte(e.Raw), &envelope)
	}
}
func BenchmarkGsonPaser(b *testing.B) {
	l := []map[string]interface{}{}
	for i := 0; i < b.N; i++ {
		m := gjson.ParseBytes(jsonData).Value().(map[string]interface{})
		l = append(l, m)
	}
	b.Log(len(l))
	b.Log("data_size:", len(jsonData))
}
