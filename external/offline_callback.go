//https://github.com/emsg/docs/wiki/RPC-Interface
package external

import (
	"encoding/json"
	"sara/config"
	"sara/utils"

	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
)

/*
输入参数：
{"sn":"889955","service":"user_message","method":"offline","params":{
	"envelope":{"id":"emsg_main@127.0.0.1_1401331382839448","ack":1,"gid":"g123","from":"usera@test.com","type":2,"ct":"1401331382839","to":"a3@test.com"},"payload":{"content":"hi all"},"vsn":"0.0.1"}
}
输出参数：
成功：
{ "sn":"789456123", "success":"true" }
失败：
{ "sn":"789456123", "success":"false", "entity":reason/exception }
*/
func OfflineCallback(packetStr string) bool {
	if !config.GetBool("enable_offline_callback", true) {
		return true
	}
	url := config.GetString("callback", "")
	if url == "" {
		return true
	}
	log4go.Debug("offline_callback_url=%s ; packet=%s", url, packetStr)
	ps := NewParams("user_message", "offline")
	m := make(map[string]interface{})
	json.Unmarshal([]byte(packetStr), &m)
	ps.Params = m
	body, _ := ps.ToJson()
	s, e := utils.PostRequest(url, "body", body)
	if e != nil {
		log4go.Error(e)
		return false
	}
	success := &rpcserver.Success{}
	e = json.Unmarshal([]byte(s), success)
	log4go.Debug("response_offline_callback: %s", success)
	if e != nil {
		log4go.Error(e)
		return false
	}
	if success.Success {
		return true
	}
	return false
}
