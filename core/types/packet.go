package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"sara/utils"
	"strings"

	"github.com/alecthomas/log4go"
	"github.com/tidwall/gjson"
)

type BasePacket struct {
	Envelope Envelope    `json:"envelope"`
	Payload  interface{} `json:"payload,omitempty"`
	Entity   *Entity     `json:"entity,omitempty"`
	Vsn      string      `json:"vsn,omitempty"`
}

type Envelope struct {
	Id   string `json:"id"`
	Jid  string `json:"jid,omitempty"`
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
	Type uint   `json:"type"`
	Ack  uint   `json:"ack,omitempty"`
	Ct   string `json:"ct,omitempty"`
	Pwd  string `json:"pwd,omitempty"`
	Gid  string `json:"gid,omitempty"`
}

//type Payload struct {
//	Attrs   map[string]interface{} `json:"attrs,omitempty"`
//	Content string                 `json:"content,omitempty"`
//}
type Entity struct {
	Result string `json:"result,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type Delay struct {
	Total   int           `json:"total"`
	Packets []*BasePacket `json:"packets,omitempty"`
}

type Packet struct {
	BasePacket
	Delay *Delay `json:"delay,omitempty"`
	raw   []byte
}

//组装离线消息，在opensession返回消息中
func (self *Packet) AddDelay(packets []*BasePacket) {
	if total := len(packets); total > 0 {
		self.Delay = &Delay{
			Total:   total,
			Packets: packets,
		}
	} else {
		self.Delay = &Delay{
			Total: 0,
		}
	}
}

func (self *Packet) ToJson() []byte {
	if self.raw != nil {
		return self.raw
	}
	if d, e := json.Marshal(self); e != nil {
		log4go.Error(e)
		return nil
	} else {
		return d
	}
}

func (self *Packet) EnvelopeIdFromToType() (id, from, to string, msgtype uint) {
	id = self.Envelope.Id
	from = self.Envelope.From
	to = self.Envelope.To
	msgtype = self.Envelope.Type
	return
}

func NewBasePacket(jsonData []byte) (*BasePacket, error) {
	packet := &BasePacket{}
	err := json.Unmarshal(jsonData, packet)
	return packet, err
}
func NewPacket(jsonData []byte) (*Packet, error) {
	packet := &Packet{}
	envelope := Envelope{}
	r := gjson.Get(string(jsonData), "envelope")
	if !r.Exists() {
		return nil, errors.New("error_packet")
	}
	envelopeRaw := r.Raw
	err := json.Unmarshal([]byte(envelopeRaw), &envelope)
	//err := json.Unmarshal(jsonData, packet)
	packet.Envelope = envelope
	packet.raw = jsonData
	return packet, err
}

func NewPacketSysNotify(id, msg string) *Packet {
	return newPacketWithEntity(id, "sys_notify", msg, MSG_TYPE_SYSTEM)
}
func NewPacketAuthFail(id, msg string) *Packet {
	return newPacketWithEntity(id, "error", msg, MSG_TYPE_OPEN_SESSION)
}
func NewPacketAck(id string) *Packet {
	//{ "envelope":{ "id":"xxxx", "from":"server_ack", "type":3, "ct":"1368410111254" }, "vsn":"0.0.1" }
	p := &Packet{}
	p.Envelope = Envelope{
		Id:   id,
		Type: MSG_TYPE_STATE,
		Ct:   fmt.Sprintf("%d", utils.Timestamp13()),
	}
	return p
}

func NewPacketAuthSuccess(id string) *Packet {
	return newPacketWithEntity(id, "ok", "", MSG_TYPE_OPEN_SESSION)
}

func newPacketWithEntity(id, result, reason string, msgtype uint) *Packet {
	p := &Packet{}
	p.Envelope = Envelope{
		Id:   id,
		Type: msgtype,
	}
	p.Entity = &Entity{Result: result}
	if reason != "" {
		p.Entity.Reason = reason
	}
	return p
}

type JID struct {
	user, domain, resource string
}

func (self *JID) GetUser() string {
	return self.user
}

func (self *JID) Equal(j string) bool {
	if self.String() == j {
		return true
	}
	return false
}
func (self *JID) EqualWithoutResource(j string) bool {
	jj, _ := NewJID(j)
	j = jj.StringWithoutResource()
	if self.StringWithoutResource() == j {
		return true
	}
	return false
}

func (self *JID) GetDomain() string {
	return self.domain
}
func (self *JID) GetResource() string {
	return self.resource
}

func (self *JID) StringWithoutResource() string {
	return fmt.Sprintf("%s@%s", self.user, self.domain)
}

func (self *JID) ToSessionid() []byte {
	return []byte(SESSION_PERFIX + self.StringWithoutResource())
}

func (self *JID) ToOfflineKey() []byte {
	return []byte(OFFLINE_PERFIX + self.StringWithoutResource())
}

func (self *JID) String() string {
	if self.resource != "" {
		return fmt.Sprintf("%s@%s/%s", self.user, self.domain, self.resource)
	} else {
		return fmt.Sprintf("%s@%s", self.user, self.domain)
	}
}
func NewJIDByUidDomain(uid, domain string) (*JID, error) {
	j := fmt.Sprintf("%s@%s", uid, domain)
	return NewJID(j)
}
func NewJID(str string) (*JID, error) {
	jid := &JID{}
	arr := strings.Split(str, "@")
	if len(arr) > 1 {
		jid.user = arr[0]
		arr2 := strings.Split(arr[1], "/")
		jid.domain = arr2[0]
		if len(arr2) == 2 {
			jid.resource = arr2[1]
		}
	} else {
		return nil, errors.New("fail format of jid :" + str)
	}
	return jid, nil
}
