package external

import (
	"sara/config"
	"testing"
)

func TestGetGroupUserList(t *testing.T) {
	url := "http://127.0.0.1:8000/request/"
	config.SetString("callback", url)
	ulist, err := GetGroupUserList("4")
	t.Log(err, ulist)
}
