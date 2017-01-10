package core

import (
	"testing"
)

func TestDecodePacketCase1(t *testing.T) {
	e := "ccc"
	buff := []byte("bbb")
	buff = append(buff, 0x1)
	buff = append(buff, e...)
	buff = append(buff, 0x1)
	buff = append(buff, e...)
	part := []byte("aa")
	pl, newPart, _ := DecodePacket(buff, part)
	t.Log(pl, newPart)
}

func TestDecodePacketCase2(t *testing.T) {
	buff := []byte("bbb")
	buff = append(buff, "ccc"...)
	buff = append(buff, "ddd"...)
	part := []byte("aa")
	pl, newPart, _ := DecodePacket(buff, part)
	t.Log(pl, newPart)
	t.Skip()
}
