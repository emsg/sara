package utils

import "testing"

func TestPost(t *testing.T) {
	res, err := PostRequest("http://www.baidu.com", "body", "{}")
	t.Log(err, res)
}
