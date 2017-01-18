package external

import (
	"encoding/json"
	"github.com/golibs/uuid"
)

type Params struct {
	Sn      string                 `json:"sn"`
	Service string                 `json:"service"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

func (self *Params) ToJson() (string, error) {
	b, e := json.Marshal(self)
	if e == nil {
		return string(b), nil
	}
	return "", e
}
func NewParams(service, method string, params ...string) *Params {
	sn := uuid.Rand().Hex()
	p := &Params{
		Sn:      sn,
		Service: service,
		Method:  method,
	}
	if t := len(params); t > 0 && t%2 == 0 {
		paramsMap := make(map[string]interface{})
		for i := 0; i < len(params); i += 2 {
			paramsMap[params[i]] = params[i+1]
		}
		p.Params = paramsMap
	}
	return p
}
