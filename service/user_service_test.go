package service

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMethodArgs(t *testing.T) {
	js := []byte(`{"username":"foo","password":"bar"}`)
	userService := &UserService{}
	rval := reflect.ValueOf(userService)
	mval := rval.MethodByName("SetUser")
	p := mval.Type().In(0)
	pv := reflect.New(p)
	pvi := pv.Interface()
	t.Log(pv.String())
	t.Log(pv)
	json.Unmarshal(js, &pvi)
	t.Log(pvi)
}
