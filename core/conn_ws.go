package core

import (
	"bytes"
	"sara/core/types"
	"time"

	"github.com/alecthomas/log4go"
	"github.com/gorilla/websocket"
)

type WsSessionConn struct {
	handler PacketHandler
	conn    *websocket.Conn
	part    []byte
	//resultCh chan *ReadPacketResult
}

func (self *WsSessionConn) SetReadTimeout(timeout time.Time) {
	self.conn.UnderlyingConn().SetReadDeadline(timeout)
	//self.conn.UnderlyingConn().SetWriteDeadline(timeout)
}

func (self *WsSessionConn) recv() {
	defer func() {
		if r := recover(); r != nil {
			log4go.Error(r)
		}
	}()
	for {
		_, p, err := self.conn.ReadMessage()
		r := &ReadPacketResult{}
		if err != nil {
			log4go.Debug("ws_recv_error => %s", err)
			r.err = err
			self.handler(r)
			return
		}
		if len(p) > 0 {
			r.packets = [][]byte{p}
			self.handler(r)
		}
	}
}
func (self *WsSessionConn) ReadPacket(handler PacketHandler) {
	//return self.resultCh
	self.handler = handler
	go self.recv()
}

func (self *WsSessionConn) WritePacket(packet []byte) (int, error) {
	//去掉尾部分隔符，websocket 协议不需要分隔符
	//r := bytes.Replace(packet, []byte{types.END_FLAG}, []byte("\n"), -1)
	r := packet[0 : len(packet)-1]
	//转义掉非 unicode 编码集的字符，javascript 中显示不了，改用对应的字符串代替
	if bytes.Compare(r, []byte{types.HEART_BEAT}) == 0 {
		r = []byte("\\02")
	} else if bytes.Compare(r, []byte{types.KILL}) == 0 {
		r = []byte("\\03")
	}
	return len(r), self.conn.WriteMessage(websocket.TextMessage, r)
}

func (self *WsSessionConn) Close() {
	self.conn.Close()
}

func NewWsSessionConn(conn *websocket.Conn) *WsSessionConn {
	sc := &WsSessionConn{
		conn: conn,
		part: make([]byte, 0),
		//resultCh: make(chan *ReadPacketResult),
	}
	//go sc.recv()
	return sc
}
