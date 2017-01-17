package config

import (
	"fmt"

	"github.com/urfave/cli"
)

var configs []ConfVo = []ConfVo{
	newConfVoKeyDef("port", 4222),
	newConfVoKeyDef("wsport", 4224),
	newConfVoKeyDef("sslport", 4333),
	newConfVoKeyDef("rpcport", 4281),
	newConfVoKeyDef("logfile", "/tmp/sara.log"),
	newConfVoKeyDef("loglevel", 3),              //0=errr, 1=warn, 2=info, 3=debug
	newConfVoKeyDef("dbaddr", "localhost:6379"), // redis
	newConfVoKeyDef("dbpool", 1000),             // redis pool size
	newConfVoKeyDef("hostname", "localhost"),    //unique,use for node to node rpc transport
	newConfVoKeyDef("dc", "dc01"),               //TODO :datacenter name; nodekey = dc:rpchost:rpcport
}

var conf map[string]ConfVo = make(map[string]ConfVo)

type ConfVo struct {
	Key      string
	Val, Def interface{}
}

func newConfVoKeyDef(key string, def interface{}) ConfVo {
	return ConfVo{
		Key: key,
		Def: def,
	}
}
func newConfVoKeyVal(key string, val interface{}) ConfVo {
	return ConfVo{
		Key: key,
		Val: val,
		Def: val,
	}
}

func init() {
	for _, c := range configs {
		fmt.Println(c.Key, c.Def)
		conf[c.Key] = c
	}
}

func SetString(k, val string) {
	conf[k] = newConfVoKeyVal(k, val)
}
func SetInt(k string, val int) {
	conf[k] = newConfVoKeyVal(k, val)
}
func SetBool(k string, val bool) {
	conf[k] = newConfVoKeyVal(k, val)
}

func GetString(k, def string) string {
	if c, ok := conf[k]; ok {
		return c.Val.(string)
	}
	return def
}
func GetInt(k string, def int) int {
	if c, ok := conf[k]; ok {
		return c.Val.(int)
	}
	return def
}
func GetBool(k string, def bool) bool {
	if c, ok := conf[k]; ok {
		return c.Val.(bool)
	}
	return def
}

func GetDef(k string) interface{} {
	if c, ok := conf[k]; ok {
		return c.Def
	}
	return nil
}

func LoadFromConsul(addr string) {

}

func LoadFromCtx(ctx *cli.Context) {
	SetInt("port", ctx.GlobalInt("port"))
	SetInt("wsport", ctx.GlobalInt("wsport"))
	SetInt("sslport", ctx.GlobalInt("sslport"))
	SetInt("rpcport", ctx.GlobalInt("rpcport"))
	SetString("logfile", ctx.GlobalString("logfile"))
	SetInt("loglevel", ctx.GlobalInt("loglevel"))
	SetString("dbaddr", ctx.GlobalString("dbaddr"))
	SetInt("dbpool", ctx.GlobalInt("dbpool"))
	SetString("hostname", ctx.GlobalString("hostname"))
	SetString("dc", ctx.GlobalString("dc"))
}
