package service

import (
	"github.com/alecthomas/log4go"
	"github.com/cc14514/go-lightrpc/rpcserver"
)

func init() {
	log4go.Info("INIT >>>>>> user_service")
}

type UserService struct{}

type UserVo struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func (self *UserService) Login(params interface{}) rpcserver.Success {
	log4go.Debug("Login.params=%s", params)
	return rpcserver.Success{
		Sn:      "111111",
		Success: true,
	}
}

func (self *UserService) GetUser(params interface{}, token rpcserver.TOKEN) rpcserver.Success {
	log4go.Debug("GetUser.params=%s", params)
	return rpcserver.Success{
		Sn:      "222222",
		Success: true,
	}
}

func (self *UserService) SetUser(vo UserVo) rpcserver.Success {
	return rpcserver.Success{
		Sn:      "333333",
		Success: true,
	}
}
