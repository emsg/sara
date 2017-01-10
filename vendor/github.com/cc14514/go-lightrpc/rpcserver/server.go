package rpcserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/alecthomas/log4go"
	"github.com/rs/cors"
	"github.com/tidwall/gjson"
)

var (
	//仅在判断参数类型时使用
	_token TOKEN
	this   *Rpcserver
)

type Rpcserver struct {
	// url , 默认 /api/
	Pattern        string
	Port           int
	ServiceMap     map[string]ServiceReg
	CheckToken     func(token TOKEN) bool
	AllowedMethods []string
}

func (self *Rpcserver) makeCors() *cors.Cors {
	log4go.Debug("StartServer port->%s ; allow_method->%s", self.Port, self.AllowedMethods)
	return cors.New(cors.Options{
		AllowedMethods: self.AllowedMethods,
	})
}

func (self *Rpcserver) StartServer() (e error) {
	this = self
	logServiceMap(self.ServiceMap)
	if len(self.AllowedMethods) == 0 {
		self.AllowedMethods = []string{"POST", "GET"}
	}
	c := self.makeCors()
	mux := http.NewServeMux()
	if self.Pattern != "" {
		mux.HandleFunc(self.Pattern, handlerFunc)
	} else {
		mux.HandleFunc("/api/", handlerFunc)
	}
	h := c.Handler(mux)
	host := fmt.Sprintf(":%d", self.Port)
	log4go.Debug("host = %s", host)
	http.ListenAndServe(host, h)
	return e
}

func logServiceMap(m map[string]ServiceReg) {
	for k, _ := range m {
		log4go.Debug("<<<< service_reg_map >>>> : %s", k)
	}
	if m == nil {
		log4go.Debug("<<<< service_reg_map_empty >>>>")
	}
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	success := &Success{Success: true}
	r.ParseForm()
	body := r.FormValue("body")
	if body == "" {
		success.Success = false
		success.Error("1000", "params of body notfound")
	} else {
		if serviceRes := gjson.Get(body, "service"); serviceRes.String() == "null" {
			success.Error("1002", "service error or notfound")
		} else if methodRes := gjson.Get(body, "method"); methodRes.String() == "null" {
			success.Error("1002", "method error or notfound")
		} else {
			s := serviceRes.String()
			//TODO check service version
			serviceReg, ok := this.ServiceMap[s]
			if ok {
				executeMethod(serviceReg, body, success)
			} else {
				success.Success = false
				success.Error("1003", "service not reg")
			}
		}
	}
	success.ResponseAsJson(w)
}

func executeMethod(serviceReg ServiceReg, body string, success *Success) {
	token := getToken(body)
	methodName := gjson.Get(body, "method").String()
	methodName = paserMethodName(methodName)
	serviceObj := serviceReg.Service
	refService := reflect.ValueOf(serviceObj)
	refMethod := refService.MethodByName(methodName)
	log4go.Debug("refService = %s, refMethod = %s", refService, refMethod)
	auth := false
	if refMethod.IsValid() {
		rmt := refMethod.Type()
		inArr := make([]reflect.Value, rmt.NumIn())
		for i := 0; i < rmt.NumIn(); i++ {
			in := rmt.In(i)
			var _token TOKEN
			log4go.Debug("in = %s", in)
			if in == reflect.TypeOf(_token) {
				log4go.Debug("TODO: AuthFilter ========>")
				inArr[i] = reflect.ValueOf(token)
				auth = true
			} else if kind := in.Kind().String(); kind == "interface" || kind == "map" {
				//interface{} 和 map[string]interface{} 就是序列化成 map 传递
				if paramsRes := gjson.Get(body, "params"); paramsRes.String() != "null" {
					target := map[string]interface{}{}
					json.Unmarshal([]byte(paramsRes.String()), &target)
					inArr[i] = reflect.ValueOf(target)
				}
			} else if kind == "string" {
				// 字符串则直接传递 json 字符串
				if paramsRes := gjson.Get(body, "params"); paramsRes.String() != "null" {
					inArr[i] = reflect.ValueOf(paramsRes.String())
				}
			} else {
				//TODO 2016-12-06 : 非常遗憾，当前版本还不能支持此功能
				// 否则反射成 in 指定的 struct 类型
				//if paramsRes := gjson.Get(body, "params"); paramsRes.String() != "null" {
				//	inVal := reflect.New(in).Interface()
				//	json.Unmarshal([]byte(paramsRes.String()), inVal)
				//	inArr[i] = reflect.ValueOf(&inVal)
				//}
				success.Success = false
				success.Error("TODO", "not support struct yet")
				return
			}
		}
		runservice := func() {
			rtn := refMethod.Call(inArr)[0].Interface().(Success)
			log4go.Debug("rtn = %s", rtn)
			success.Sn = rtn.Sn
			success.Success = rtn.Success
			success.Entity = rtn.Entity
		}
		if auth {
			if this.CheckToken(token) {
				runservice()
			} else {
				success.Error("1003", fmt.Sprintf("error token"))
			}
		} else {
			runservice()
		}
	} else {
		success.Error("1002", fmt.Sprintf("method notfond ;;; %s", methodName))
	}
}

func paserMethodName(s string) string {
	b0 := s[0]
	s0 := string(b0)
	s0 = strings.ToUpper(s0)
	sn := s[1:len(s)]
	s = s0 + sn
	return s
}

func getToken(body string) TOKEN {
	var token TOKEN
	if tokenRes := gjson.Get(body, "token"); tokenRes.String() != "null" {
		token = TOKEN(tokenRes.String())
	}
	return token
}
