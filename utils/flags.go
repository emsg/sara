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
		Usage: "thrift rpc port",
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
	HostnameFlag = cli.StringFlag{
		Name:  "hostname",
		Usage: "unique,use for node to node transport",
		Value: config.GetDef("hostname").(string),
	}
	DcFlag = cli.StringFlag{
		Name:  "dc",
		Usage: "TODO:datacenter name; nodekey = dc:rpchost:rpcport",
		Value: config.GetDef("dc").(string),
	}
	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "write 'pprof' info to /tmp/sara_cpu.out and /tmp/sara_mem.out",
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
		DBPoolFlag,
		HostnameFlag,
		DcFlag,
		DebugFlag,
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
		Addr,
		HbType,
		ConnType,
	}
}
