package sararpc

import (
	"emsg/emsg_inf_push"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/alecthomas/log4go"
)

type RPCServer struct {
	Addr             string
	transportFactory thrift.TTransportFactory
	protocolFactory  *thrift.TBinaryProtocolFactory
	serverSocket     *thrift.TServerSocket
	stop             chan struct{}
	dataChannel      DataChannel
}

type emsgInfPush struct {
	dataChannel DataChannel
}

func (self *emsgInfPush) Process(licence string, sn string, content string) (r string, err error) {
	log4go.Debug("rpc_process(%s , %s , %s)", licence, sn, content)
	//TODO verify licence
	self.dataChannel.Publish(self.dataChannel.GetChannel(), content)
	return "success", nil
}

func (self *emsgInfPush) ProcessBatch(licence string, sn string, contents []string) (r string, err error) {
	//TODO verify licence
	for _, content := range contents {
		self.dataChannel.Publish(self.dataChannel.GetChannel(), content)
	}
	return "success", nil
}
func newEmsgInfoPush(dataChannel DataChannel) *emsgInfPush {
	return &emsgInfPush{
		dataChannel: dataChannel,
	}
}

func (self *RPCServer) Start() {
	handler := newEmsgInfoPush(self.dataChannel)
	processor := emsg_inf_push.NewEmsgInfPushProcessor(handler)
	server := thrift.NewTSimpleServer4(processor, self.serverSocket, self.transportFactory, self.protocolFactory)
	log4go.Info("RPCServer listener on  [%s]", self.Addr)
	server.Serve()
}

func NewRPCServer(addr string, dataChannel DataChannel) (*RPCServer, error) {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	serverSocket, err := thrift.NewTServerSocket(addr)
	if err != nil {
		return nil, err
	}
	return &RPCServer{
		Addr:             addr,
		transportFactory: transportFactory,
		protocolFactory:  protocolFactory,
		serverSocket:     serverSocket,
		dataChannel:      dataChannel,
	}, nil
}
