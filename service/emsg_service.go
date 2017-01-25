package service

import (
	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
)

type EmsgGroupService struct{}

func (self *EmsgGroupService) Reload(params interface{}, token rpcserver.TOKEN) rpcserver.Success {
	log4go.Debug("Reload.params=%s", params)
	return rpcserver.Success{Success: true}
}
