package node

import (
	"errors"
	"sara/core"
	"sara/core/types"
	"sara/saradb"
	"sara/sararpc"

	"github.com/alecthomas/log4go"
)

type routerItem struct {
	channel, sid string
	packet       *types.Packet
	signal       byte
}

func (self *routerItem) vals() (channel, sid string, packet *types.Packet, signal byte) {
	channel, sid, packet, signal = self.channel, self.sid, self.packet, self.signal
	return
}

type Router struct {
	db           saradb.Database
	dataChannel  sararpc.DataChannel
	routerItemCh chan *routerItem
}

func (self *Router) Route(channel, sid string, packet *types.Packet, signal ...byte) error {
	item := &routerItem{
		channel: channel,
		sid:     sid,
		packet:  packet,
	}
	if signal != nil {
		item.signal = signal[0]
	} else {
		item.signal = 0
	}
	select {
	case self.routerItemCh <- item:
		return nil
	default:
		return errors.New("load_over")
	}
}

func (self *Router) worker() {
	for {
		item, ok := <-self.routerItemCh
		if !ok {
			return
		}
		channel, sid, packet, signal := item.vals()
		if channel == "" || self.dataChannel.GetChannel() == channel {
			if session, ok := fetchSession(sid); ok {
				if signal != 0 {
					log4go.Debug("👮 node.route_signal -> %s", signal)
					if signal == types.KILL {
						session.Kill()
					}
				} else {
					log4go.Debug("👮 node.route_packet-> %s", packet.ToJson())
					session.RoutePacket(packet)
				}
			}
		} else {
			if err := self.dataChannel.Publish(channel, string(packet.ToJson())); err != nil {
				log4go.Error("❌  dataChannel.Publish_err: %s", err)
				core.StorePacket(self.db, packet)
			}
		}
	}
}

func newRouter(db saradb.Database, dataChannel sararpc.DataChannel) *Router {
	router := &Router{
		db:           db,
		dataChannel:  dataChannel,
		routerItemCh: make(chan *routerItem, 100000),
	}
	for i := 0; i < 1000; i++ {
		go router.worker()
	}
	return router
}
