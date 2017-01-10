package service

import "github.com/cc14514/go-lightrpc/rpcserver"

var ServiceRegMap map[string]rpcserver.ServiceReg = make(map[string]rpcserver.ServiceReg)

func genServiceReg(namespace, version string, service interface{}) {
	ServiceRegMap[namespace] = rpcserver.ServiceReg{
		Namespace: namespace,
		Version:   version,
		Service:   service,
	}
}

func init() {
	vsn := "0.0.1"
	genServiceReg("user", vsn, &UserService{})
}
