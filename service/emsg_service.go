package service

import (
	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
	"github.com/golibs/uuid"
)

func init() {
	log4go.Info("INIT >>>>>> user_service")
}

type EmsgGroupService struct{}

func (self *EmsgGroupService) Reload(params interface{}, token rpcserver.TOKEN) rpcserver.Success {
	sn := uuid.Rand().Hex()
	log4go.Debug("Reload.params=%s", params)
	return rpcserver.Success{
		Sn:      sn,
		Success: true,
	}
}
