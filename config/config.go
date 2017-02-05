package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alecthomas/log4go"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli"
)

var configs []ConfVo = []ConfVo{
	newConfVoKeyDef("port", 4222),
	newConfVoKeyDef("wsport", 4224),
	newConfVoKeyDef("tlsport", 4333),
	newConfVoKeyDef("rpcport", 4280),
	newConfVoKeyDef("logfile", "/tmp/sara.log"),
	newConfVoKeyDef("loglevel", 3),                //0=errr, 1=warn, 2=info, 3=debug
	newConfVoKeyDef("dbaddr", "localhost:6379"),   // redis
	newConfVoKeyDef("dbpool", 1000),               // redis pool size
	newConfVoKeyDef("nodeaddr", "localhost:4281"), //unique,use for node to node rpc transport
	newConfVoKeyDef("callback", ""),               //callbackurl,for auth、offline notify、fetch group users
	newConfVoKeyDef("dc", "dc01"),                 //TODO :datacenter name; nodekey = dc:hostname
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
		//fmt.Println(c.Key, c.Def)
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

func LoadFromConf(c *cli.Context) {
	addr := c.GlobalString("config")
	buf, err := ioutil.ReadFile(addr)
	if err != nil {
		//log4go.Error(err)
		return
	}
	j := string(buf)
	log4go.Debug("config= %s", j)
	if r := gjson.Get(j, "nodeid"); r.Exists() && r.String() != "" {
		SetString("nodeid", r.String())
	} else if nid := GetString("nodeid", ""); nid == "" {
		fmt.Println("⚠️  nodeid can not empty")
		os.Exit(0)
	}
	if r := gjson.Get(j, "port"); r.Exists() {
		SetInt("port", int(r.Int()))
	}
	if r := gjson.Get(j, "wsport"); r.Exists() {
		SetInt("wsport", int(r.Int()))
	}
	if r := gjson.Get(j, "tlsport"); r.Exists() {
		SetInt("tlsport", int(r.Int()))
	}
	if r := gjson.Get(j, "wssport"); r.Exists() {
		SetInt("wssport", int(r.Int()))
	}
	if r := gjson.Get(j, "rpcport"); r.Exists() {
		SetInt("rpcport", int(r.Int()))
	}
	if r := gjson.Get(j, "accesstoken"); r.Exists() {
		SetString("accesstoken", r.String())
	}
	if r := gjson.Get(j, "logfile"); r.Exists() {
		SetString("logfile", r.String())
	}
	if r := gjson.Get(j, "loglevel"); r.Exists() {
		SetInt("loglevel", int(r.Int()))
	}
	if r := gjson.Get(j, "dbaddr"); r.Exists() {
		SetString("dbaddr", r.String())
	}
	if r := gjson.Get(j, "dbpool"); r.Exists() {
		SetInt("dbpool", int(r.Int()))
	}
	if r := gjson.Get(j, "nodeaddr"); r.Exists() {
		SetString("nodeaddr", r.String())
	}
	if r := gjson.Get(j, "callback"); r.Exists() {
		SetString("callback", r.String())
	}
	if r := gjson.Get(j, "dc"); r.Exists() {
		SetString("dc", r.String())
	}
	if r := gjson.Get(j, "certfile"); r.Exists() {
		SetString("certfile", r.String())
	}
	if r := gjson.Get(j, "keyfile"); r.Exists() {
		SetString("keyfile", r.String())
	}
	if r := gjson.Get(j, "enable_auth"); r.Exists() {
		SetBool("enable_auth", r.Bool())
	}
	if r := gjson.Get(j, "enable_offline_callback"); r.Exists() {
		SetBool("enable_offline_callback", r.Bool())
	}

	if r := gjson.Get(j, "enable_tcp"); r.Exists() {
		SetBool("enable_tcp", r.Bool())
	}
	if r := gjson.Get(j, "enable_tls"); r.Exists() {
		SetBool("enable_tls", r.Bool())
	}
	if r := gjson.Get(j, "enable_ws"); r.Exists() {
		SetBool("enable_ws", r.Bool())
	}
	if r := gjson.Get(j, "enable_wss"); r.Exists() {
		SetBool("enable_wss", r.Bool())
	}

}

func LoadFromCtx(ctx *cli.Context) {
	SetInt("port", ctx.GlobalInt("port"))
	SetInt("wsport", ctx.GlobalInt("wsport"))
	SetInt("tlsport", ctx.GlobalInt("tlsport"))
	SetInt("rpcport", ctx.GlobalInt("rpcport"))
	SetString("logfile", ctx.GlobalString("logfile"))
	SetInt("loglevel", ctx.GlobalInt("loglevel"))
	SetString("dbaddr", ctx.GlobalString("dbaddr"))
	SetInt("dbpool", ctx.GlobalInt("dbpool"))
	SetString("nodeaddr", ctx.GlobalString("nodeaddr"))
	SetString("callback", ctx.GlobalString("callback"))
	SetString("dc", ctx.GlobalString("dc"))
	SetString("nodeid", ctx.GlobalString("nodeid"))
}
func Load(ctx *cli.Context) {
	//LoadFromCtx(ctx)
	LoadFromConf(ctx)
}

var Template string = `
{
    "port": 4222,
    "wsport": 4224,
    "tlsport": 4333,
    "wssport": 4334,
    "rpcport": 4280,
    "accesstoken":"http-rpc access token",
    "nodeid":"n01",
    "dbaddr": "localhost:6379",
    "dbpool":100,
    "callback":"",
    "nodeaddr": "localhost:4281",
    "logfile":"/tmp/sara.log",
    "loglevel":3,
    "dc":"dc01",
    "certfile":"/etc/sara/server.pem",
    "keyfile":"/etc/sara/server.key",
    "enable_tcp":true,
    "enable_tls":true,
    "enable_ws":true,
    "enable_wss":true,
    "enable_auth":true,
    "enable_offline_callback":true
}
`
