package service

import (
	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
	"github.com/urfave/cli"
)

func StartRPC(ctx *cli.Context) error {
	rpcport := ctx.GlobalInt("rpcport")
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
