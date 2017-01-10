package core

import (
	"bytes"
	"encoding/json"
	"sara/core/types"

	"github.com/alecthomas/log4go"
)

func DecodePacket(data []byte, part []byte) ([][]byte, []byte, error) {
	var (
		new_part []byte
		packets  [][]byte
	)
	//log4go.Debug("1🐍 >>>> %b", data)
	//去掉 buff 结尾的占位符
	if last := bytes.IndexByte(data, 0x0); last > 0 {
		data = data[:last]
	}
	//log4go.Debug("2🐍 >>>> %b", data)

	dataList := bytes.Split(data, []byte{types.END_FLAG})
	//log4go.Debug("split.len -> %d", len(dataList))
	lastByte := data[len(data)-1]
	size := len(dataList) - 1
	if size > 0 {
		for i, packet := range dataList {
			if len(packet) > 0 {
				if i == 0 && len(part) > 0 {
					packet = append(part, packet...)
					packets = append(packets, packet)
				} else if i == size && lastByte != types.END_FLAG {
					new_part = packet
				} else {
					packets = append(packets, packet)
				}
			}
		}
	} else {
		packet := dataList[0]
		if len(part) > 0 {
			packet = append(part, packet...)
		}
		if lastByte != types.END_FLAG {
			new_part = packet
		} else {
			packets = append(packets, packet)
		}
	}
	return packets, new_part, nil
}

func UnmarshalPacket(data []byte) (*types.Packet, error) {
	log4go.Debug("⚙️ unmarshal-> %s", data)
	packet := &types.Packet{}
	err := json.Unmarshal(data, packet)
	return packet, err
}
