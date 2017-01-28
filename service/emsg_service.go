package service

import (
	"fmt"
	"sara/core/types"
	"sara/node"

	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
)

/*
接口文档
https://github.com/emsg/docs/wiki/RPC
*/

type EmsgGroupService struct{ node *node.Node }

// 1 重新加载群成员
func (self *EmsgGroupService) Reload(params interface{}, token rpcserver.TOKEN) rpcserver.Success {
	log4go.Debug("Reload.params=%s", params)
	p := params.(map[string]interface{})
	gid, gid_ok := p["gid"]
	if !gid_ok || gid == "" {
		return SuccessFalse("gid_not_be_empty")
	}
	domain, domain_ok := p["domain"]
	if !domain_ok || domain == "" {
		return SuccessFalse("domain_not_be_empty")
	}
	key := fmt.Sprintf("group_%s@%s", gid, domain)
	self.node.GetDB().Delete([]byte(key))
	log4go.Debug(key)
	return rpcserver.Success{Success: true}
}

type EmsgSessionService struct{ node *node.Node }

// 2 当前在线用户总数
func (self *EmsgSessionService) Counter(params interface{}, token rpcserver.TOKEN) rpcserver.Success {
	db := self.node.GetDB()
	nodeids, err := db.GetByIdx([]byte("sara"))
	if err != nil || len(nodeids) < 1 {
		return SuccessFalse("notfound")
	}
	entity := make(map[string]interface{})
	counters := make([]map[string]interface{}, 0)
	for _, nodeid := range nodeids {
		if count, err := db.CountByIdx(nodeid); err == nil {
			c := make(map[string]interface{})
			c["node"] = string(nodeid)
			c["count"] = count
			counters = append(counters, c)
		}
	}
	entity["counters"] = counters
	return rpcserver.Success{Success: true, Entity: entity}
}

//3 判断用户是否在线
func (self *EmsgSessionService) Isonline(params interface{}, token rpcserver.TOKEN) rpcserver.Success {
	db := self.node.GetDB()
	p := params.(map[string]interface{})
	if jid, ok := p["jid"]; ok {
		if j, err := types.NewJID(jid.(string)); err == nil {
			sid := j.ToSessionid()
			if _, e := db.Get(sid); e == nil {
				return rpcserver.Success{Success: true}
			}
		}
	}
	return rpcserver.Success{Success: false}
}
