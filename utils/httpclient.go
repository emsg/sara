package utils

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/alecthomas/log4go"
)

//func PostRequest(urlStr string, params map[string]string) (string, error) {
func PostRequest(urlStr string, params ...string) (string, error) {
	log4go.Debug(params)
	vals := url.Values{}
	for i := 0; i < len(params); i += 2 {
		k, v := params[i], params[i+1]
		vals.Set(k, v)
	}
	body := strings.NewReader(vals.Encode())
	client := &http.Client{}
	request, _ := http.NewRequest("POST", urlStr, body)

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Cookie", "name=body")
	response, err1 := client.Do(request)
	if err1 != nil {
		return "", err1
	}
	defer response.Body.Close()
	data, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		return "", err2
	}
	return string(data), nil
}
