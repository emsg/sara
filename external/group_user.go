//https://github.com/emsg/docs/wiki/RPC-Interface
package external

import (
	"encoding/json"
	"errors"
	"sara/config"
	"sara/utils"

	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
	"github.com/tidwall/gjson"
)

/*
Input:
{ "sn": "9966554", "service": "emsg_group", "method": "get_user_list", "params": { "gid":"群ID" } }
Output:
成功：
{ "sn":"9966554", "success":"true", "entity":{"users":["user1","user2","user3",...]} }
失败：
{ "sn":"789456123", "success":"false", "entity":reason/exception }
*/
func GetGroupUserList(gid string) (userlist []string, err error) {
	url := config.GetString("callback", "")
	if url == "" {
		return nil, errors.New("callback_url_not_define")
	}
	body, body_err := NewParams("emsg_group", "get_user_list", "gid", gid).ToJson()
	if body_err != nil {
		log4go.Error(body_err)
		err = body_err
		return
	}
	s, e := utils.PostRequest(url, "body", body)
	if e != nil {
		log4go.Error(e)
		err = e
	}
	success := &rpcserver.Success{}
	e = json.Unmarshal([]byte(s), success)
	log4go.Debug("response_group_user_list: %s", success)
	if e != nil {
		log4go.Error(e)
		err = e
	}
	if success.Success {
		for _, res := range gjson.Get(s, "entity.users").Array() {
			userlist = append(userlist, res.String())
		}
	} else {
		err = errors.New(gjson.Get(s, "entity").String())
	}
	return
}
