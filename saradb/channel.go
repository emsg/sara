package saradb

import (
	"github.com/alecthomas/log4go"
	"github.com/mediocregopher/radix.v2/pubsub"
	"github.com/mediocregopher/radix.v2/redis"
)

//TODO 这个对象是需要连接池的
//暂时简单实现单通道,其实单通道性能也不低
type NodeChannel struct {
	id   string
	size int
	sub  *pubsub.SubClient
	pub  *redis.Client
}

func (self *NodeChannel) GetChannel() string {
	// TODO 动态负载分配,根据 size 生成 n 个 channel，平均分配 session
	return self.id
}

func (self *NodeChannel) Publish(channel, message string) error {
	r := self.pub.Cmd("PUBLISH", channel, message)
	log4go.Debug("🔫  c=%s,m=%s,r=%s,e=%s", channel, message, r, r.Err)
	return nil
}

func (self *NodeChannel) Subscribe(handler SubHandler) {
	self.sub.Subscribe(self.id)
	//subChan := make(chan *pubsub.SubResp)
	go func() {
		for {
			r := self.sub.Receive()
			handler(r.Message)
		}
	}()
}

func newChannel(name string, sub, pub *redis.Client) *NodeChannel {
	nc := &NodeChannel{
		id:  name,
		sub: pubsub.NewSubClient(sub),
		pub: pub,
	}
	return nc
}
