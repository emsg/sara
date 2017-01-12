package utils

import "github.com/urfave/cli"

var (
	ListenPortFlag = cli.IntFlag{
		Name:  "port",
		Usage: "Network listening port",
		Value: 4222,
	}
	ListenWSPortFlag = cli.IntFlag{
		Name:  "wsport",
		Usage: "Network listening websocket port",
		Value: 4224,
	}
	ListenSSLPortFlag = cli.IntFlag{
		Name:  "sslport",
		Usage: "TODO:Network listening port",
		Value: 4333,
	}
	ListenRPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "TODO:Network listening port",
		Value: 4280,
	}
	LogfileFlag = cli.StringFlag{
		Name:  "logfile",
		Usage: "log file path",
	}
	LogLevelFlag = cli.IntFlag{
		Name:  "loglevel",
		Usage: "0=errr, 1=warn, 2=info, 3=debug",
		Value: 3,
	}
	DBAddrFlag = cli.StringFlag{
		Name:  "dbaddr",
		Usage: "redis addr,format as ip:port ",
		Value: "localhost:6379",
	}
	NodeFlag = cli.StringFlag{
		Name:  "node",
		Usage: "node name , unique  (default: uuid.hex)",
		Value: "",
	}
)

func InitFlags() []cli.Flag {
	return []cli.Flag{
		ListenPortFlag,
		ListenWSPortFlag,
		ListenSSLPortFlag,
		ListenRPCPortFlag,
		LogfileFlag,
		LogLevelFlag,
		DBAddrFlag,
		NodeFlag,
	}
}

var (
	Total = cli.IntFlag{
		Name:  "total,t",
		Usage: "execute 'ulimit -n' to fetch the max value",
		Value: 1024,
	}
	Addr = cli.StringFlag{
		Name:  "addr,a",
		Usage: "host:port",
		Value: "localhost:4222",
	}
	ConnType = cli.IntFlag{
		Name:  "conn_type",
		Usage: "0 tcp,1 ws,2 ssl,3 wss",
		Value: 0,
	}
)

func InitFlagsForTestOfMakeConn() []cli.Flag {
	return []cli.Flag{
		Total,
		Addr,
		ConnType,
	}
}
