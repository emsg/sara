package sararpc

import (
	"emsg/emsg_inf_push"
	"git.apache.org/thrift.git/lib/go/thrift"
)

//每个通道的连接池阀值大小
const PPSize int = 100

type RPCClient struct {
	transportFactory thrift.TTransportFactory
	protocolFactory  *thrift.TBinaryProtocolFactory
	clientPool       map[string]chan *emsg_inf_push.EmsgInfPushClient
	poolSize         int
}

func (self *RPCClient) newClient(addr string) (*emsg_inf_push.EmsgInfPushClient, error) {
	transport, err := thrift.NewTSocket(addr)
	if err != nil {
		return nil, err
	}
	if err = transport.Open(); err != nil {
		return nil, err
	}
	useTransport := self.transportFactory.GetTransport(transport)
	client := emsg_inf_push.NewEmsgInfPushClientFactory(useTransport, self.protocolFactory)
	return client, nil
}

func (self *RPCClient) putClient(addr string, c *emsg_inf_push.EmsgInfPushClient) {
	var queue chan *emsg_inf_push.EmsgInfPushClient
	if clientCh, ok := self.clientPool[addr]; ok {
		queue = clientCh
	} else {
		queue = make(chan *emsg_inf_push.EmsgInfPushClient, self.poolSize)
		self.clientPool[addr] = queue
	}
	select {
	case queue <- c:
		//log4go.Debug("put: %s", addr)
	default:
		//log4go.Debug("close: %s", addr)
		c.Transport.Close()
	}
}

func (self *RPCClient) getClient(addr string) (*emsg_inf_push.EmsgInfPushClient, error) {
	var queue chan *emsg_inf_push.EmsgInfPushClient
	if clientCh, ok := self.clientPool[addr]; ok {
		queue = clientCh
	} else {
		queue = make(chan *emsg_inf_push.EmsgInfPushClient, self.poolSize)
		self.clientPool[addr] = queue
	}
	select {
	case client := <-queue:
		//log4go.Debug("pool")
		return client, nil
	default:
		//log4go.Debug("new")
		return self.newClient(addr)
	}
}

func (self *RPCClient) Call(addr, licence, sn, content string) (string, error) {
	client, err := self.getClient(addr)
	if err != nil {
		return "", err
	}
	if rtn, e := client.Process(licence, sn, content); e != nil {
		return rtn, e
	} else {
		self.putClient(addr, client)
		return rtn, nil
	}
}

func NewRPCClient() *RPCClient {
	rpcclient := &RPCClient{poolSize: PPSize, clientPool: make(map[string]chan *emsg_inf_push.EmsgInfPushClient)}
	rpcclient.transportFactory = thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	rpcclient.protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	return rpcclient
}
