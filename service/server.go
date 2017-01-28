package service

import (
	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
	"sara/config"
	"sara/node"
)

var _node *node.Node

func SuccessFalse(reason string) rpcserver.Success {
	entity := make(map[string]interface{})
	entity["reason"] = reason
	return rpcserver.Success{
		Success: false,
		Entity:  entity,
	}
}

func getNode() *node.Node {
	return _node
}

func StartRPC(node *node.Node) error {
	_node = node
	rpcport := config.GetInt("rpcport", 4280)
	log4go.Info("http-rpc start on [0.0.0.0:%d]", rpcport)
	go func(rpcport int) {
		RegService()
		rs := &rpcserver.Rpcserver{
			Port:       rpcport,
			ServiceMap: ServiceRegMap,
			// 校验请求中的 TOKEN 是否正确，根据不同的业务需求，会有不同实现
			CheckToken: func(token rpcserver.TOKEN) bool {
				if accesstoken := config.GetString("accesstoken", ""); accesstoken == "" {
					log4go.Warn("⚠️  http-rpc free accese now, please set accesstoken.")
					return true
				} else if accesstoken == string(token) {
					return true
				}
				log4go.Debug("error_token: %s", token)
				return false
			},
		}
		rs.StartServer()
	}(rpcport)
	return nil
}
