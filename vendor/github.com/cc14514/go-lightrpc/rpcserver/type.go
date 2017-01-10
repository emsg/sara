package rpcserver

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ServiceReg struct {
	Namespace string
	Version   string
	Service   interface{}
}

type TOKEN string

type AppRequest struct {
	Sn      string      `json:"sn"`
	Service string      `json:"service"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Token   TOKEN       `json:"token,omitempty"`
}

type ReasonEntity struct {
	ErrCode string      `json:"errCode,omitempty"`
	Reason  interface{} `json:"reason"`
}

type Success struct {
	Sn      string      `json:"sn"`
	Success bool        `json:"success"`
	Entity  interface{} `json:"entity,omitempty"`
}

func (self *Success) Error(errCode string, reason interface{}) {
	self.Success = false
	self.Entity = ReasonEntity{
		ErrCode: errCode,
		Reason:  reason,
	}
}

// 把 Success 对象序列化成 JSON string 写到 response 中
func (self *Success) ResponseAsJson(wp http.ResponseWriter) {
	wp.Header().Set("Content-Type", "application/json")
	j, e := json.Marshal(self)
	if e != nil {
		self.Success = false
		self.Entity = ReasonEntity{
			ErrCode: "1000",
			Reason:  fmt.Sprintf("%s", e),
		}
		j, _ = json.Marshal(self)
	}
	wp.Write(j)
}
