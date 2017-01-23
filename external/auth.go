//https://github.com/emsg/docs/wiki/RPC-Interface
package external

import (
	"encoding/json"
	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
	"sara/config"
	"sara/utils"
)

/*
输入参数：其中 params 参数为离线时的数据包，需要后台解析后做相关处理；
{ "sn": "789456123", "service": "emsg_auth", "method": "auth", "params": { "uid":"xxx", "token":"需要认证的身份证明" } }
输出参数：
成功：
{ "sn":"789456123", "success":"true" }
失败：
{ "sn":"789456123", "success":"false", "entity":reason/exception }
*/
func Auth(uid, pwd string) bool {
	url := config.GetString("callback", "")
	if url == "" {
		return true
	}
	log4go.Debug("auth_callback_url=%s ; uid=%s,pwd=%s", url, uid, pwd)
	body, body_err := NewParams("emsg_auth", "auth", "uid", uid, "token", pwd).ToJson()
	if body_err != nil {
		log4go.Error(body_err)
		return false
	}
	s, e := utils.PostRequest(url, "body", body)
	if e != nil {
		log4go.Error(e)
		return false
	}
	success := &rpcserver.Success{}
	e = json.Unmarshal([]byte(s), success)
	log4go.Debug("response_auth: %s", success)
	if e != nil {
		log4go.Error(e)
		return false
	}
	if success.Success {
		return true
	}
	return false
}
