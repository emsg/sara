package utils

import "github.com/urfave/cli"
import "sara/config"

var (
	ListenPortFlag = cli.IntFlag{
		Name:  "port",
		Usage: "Network listening port",
		Value: config.GetDef("port").(int),
	}
	ListenWSPortFlag = cli.IntFlag{
		Name:  "wsport",
		Usage: "Network listening websocket port",
		Value: config.GetDef("wsport").(int),
	}
	ListenSSLPortFlag = cli.IntFlag{
		Name:  "sslport",
		Usage: "TODO:Network listening port",
		Value: config.GetDef("sslport").(int),
	}
	ListenRPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "Network listening http-rpc port",
		Value: config.GetDef("rpcport").(int),
	}
	LogfileFlag = cli.StringFlag{
		Name:  "logfile",
		Usage: "log file path",
		Value: config.GetDef("logfile").(string),
	}
	LogLevelFlag = cli.IntFlag{
		Name:  "loglevel",
		Usage: "0=errr, 1=warn, 2=info, 3=debug",
		Value: config.GetDef("loglevel").(int),
	}
	DBAddrFlag = cli.StringFlag{
		Name:  "dbaddr",
		Usage: "redis addr,format as ip:port ",
		Value: config.GetDef("dbaddr").(string),
	}
	DBPoolFlag = cli.IntFlag{
		Name:  "dbpool",
		Usage: "redis pool size",
		Value: config.GetDef("dbpool").(int),
	}
	NodeaddrFlag = cli.StringFlag{
		Name:  "nodeaddr,n",
		Usage: "unique,use for node to node transport",
		Value: config.GetDef("nodeaddr").(string),
	}
	DcFlag = cli.StringFlag{
		Name:  "dc",
		Usage: "TODO:datacenter name; nodekey = dc:nodeaddr",
		Value: config.GetDef("dc").(string),
	}
	CallbackFlag = cli.StringFlag{
		Name:  "callback",
		Usage: "callbackurl,for auth、offline notify、fetch group users",
	}
	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "write 'pprof' info to /tmp/sara_cpu.out and /tmp/sara_mem.out",
	}
	ConfigFlag = cli.StringFlag{
		Name:  "config,c",
		Usage: "cmd-line first, config second ",
		Value: "/etc/sara/conf.json",
	}
	NodeidFlag = cli.StringFlag{
		Name:  "nodeid",
		Usage: "unique and not empty",
	}
)

func InitFlags() []cli.Flag {
	return []cli.Flag{
		/*
			ListenPortFlag,
			ListenWSPortFlag,
			ListenSSLPortFlag,
			ListenRPCPortFlag,
			LogfileFlag,
			LogLevelFlag,
			DBAddrFlag,
			DBPoolFlag,
			NodeidFlag,
			NodeaddrFlag,
			DcFlag,
			CallbackFlag,
		*/
		DebugFlag,
		ConfigFlag,
	}
}

var (
	Total = cli.IntFlag{
		Name:  "total,t",
		Usage: "execute 'ulimit -n' to fetch the max value",
		Value: 1024,
	}
	Laddr = cli.StringFlag{
		Name:  "laddr,l",
		Usage: "source ip",
		Value: "",
	}
	Addr = cli.StringFlag{
		Name:  "raddr,a",
		Usage: "target ip:port ",
		Value: "localhost:4222",
	}
	HbType = cli.IntFlag{
		Name:  "heartbeat,b",
		Usage: "second",
		Value: 50,
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
		Laddr,
		Addr,
		HbType,
		ConnType,
	}
}
