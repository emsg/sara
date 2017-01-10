package core

import (
	"errors"
	"net"
	"time"

	"github.com/alecthomas/log4go"
)

type TcpSessionConn struct {
	conn net.Conn
}

func (self *TcpSessionConn) SetReadTimeout(timeout time.Time) {
	self.conn.SetReadDeadline(timeout)
}

func (self *TcpSessionConn) ReadPacket(part []byte) (packetList [][]byte, newPart []byte, e error) {
	buff := make([]byte, 256)
	if _, err := self.conn.Read(buff); err != nil {
		log4go.Error("❌❌  err=%s , et=%t", err.Error(), err)
		switch err.(type) {
		case *net.OpError:
			e = errors.New("TIMEOUT")
		default:
			e = errors.New("EOF")
		}
	} else {
		packetList, newPart, _ = DecodePacket(buff, part)
	}
	return
}

func (self *TcpSessionConn) WritePacket(packet []byte) (int, error) {
	return self.conn.Write(packet)
}

func (self *TcpSessionConn) Close() {
	self.conn.Close()
}

func NewTcpSessionConn(conn net.Conn) *TcpSessionConn {
	sc := &TcpSessionConn{}
	sc.conn = conn
	return sc
}
