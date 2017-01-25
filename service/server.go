package service

import (
	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
	"sara/config"
	"sara/node"
)

func StartRPC(node *node.Node) error {
	rpcport := config.GetInt("rpcport", 4280)
	log4go.Info("http-rpc start on [0.0.0.0:%d]", rpcport)
	go func(rpcport int) {
		rs := &rpcserver.Rpcserver{
			Port:       rpcport,
			ServiceMap: ServiceRegMap,
			// 校验请求中的 TOKEN 是否正确，根据不同的业务需求，会有不同实现
			CheckToken: func(token rpcserver.TOKEN) bool {
				log4go.Debug("TODO: Auth token = %s", token)
				return true
			},
		}
		log4go.Info("rpcserver is running at : %s", rpcport)
		rs.StartServer()
	}(rpcport)
	return nil
}
