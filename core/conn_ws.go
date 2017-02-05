package core

import (
	"bytes"
	"sara/core/types"
	"time"

	"github.com/alecthomas/log4go"
	"github.com/gorilla/websocket"
)

type WsSessionConn struct {
	conn *websocket.Conn
}

func (self *WsSessionConn) SetReadTimeout(timeout time.Time) {
	self.conn.UnderlyingConn().SetReadDeadline(timeout)
}

func (self *WsSessionConn) ReadPacket(part []byte) ([][]byte, []byte, error) {
	if _, p, err := self.conn.ReadMessage(); err != nil {
		// 其实 eof/timeout/others 都无所谓，session 单独处理 eof
		// 只是想给一个通知,简单起见，这个可以没有
		log4go.Debug("🌍  --> err = %s , %v", err.Error(), err)
		return nil, nil, err
	} else {
		return [][]byte{p}, part, nil
	}
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
	return &WsSessionConn{conn: conn}
}
