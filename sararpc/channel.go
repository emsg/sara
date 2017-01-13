package sararpc

import "github.com/golibs/uuid"

const LICENCE string = "ABECDD3C-B737-42BA-8F49-92A3267BB822"

type RPCDataChannel struct {
	id        string //format like 'host:port', use for rpc call
	dataCh    chan string
	rpcclient *RPCClient
}

func (self *RPCDataChannel) GetChannel() string {
	return self.id
}

func (self *RPCDataChannel) Publish(channel string, message string) error {
	if channel == self.GetChannel() {
		self.dataCh <- message
		return nil
	} else {
		sn := uuid.Rand().Hex()
		_, e := self.rpcclient.Call(channel, LICENCE, sn, message)
		return e
	}
}

func (self *RPCDataChannel) Subscribe(handler SubHandler) {
	go func() {
		for {
			message := <-self.dataCh
			handler(message)
		}
	}()
}

func NewRPCDataChannel(rpcserverAddr string, bufSize int) *RPCDataChannel {
	return &RPCDataChannel{
		id:        rpcserverAddr,
		dataCh:    make(chan string, bufSize),
		rpcclient: NewRPCClient(),
	}
}
