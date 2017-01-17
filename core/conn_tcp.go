package core

import (
	"errors"
	"github.com/alecthomas/log4go"
	"net"
	"time"
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
		switch err.(type) {
		case *net.OpError:
			ne := err.(*net.OpError)
			if ne.Timeout() {
				// socket timeout
				e = errors.New("TIMEOUT")
			} else {
				// server socket close or others
				e = errors.New(ne.Error())
			}
		default:
			// EOF, normal
			e = err
		}
	} else {
		packetList, newPart, _ = DecodePacket(buff, part)
	}
	return
}

func (self *TcpSessionConn) WritePacket(packet []byte) (i int, e error) {
	defer func() {
		if err := recover(); err != nil {
			log4go.Debug(err)
			i, e = 0, errors.New(err.(string))
		}
	}()
	i, e = self.conn.Write(packet)
	return
}

func (self *TcpSessionConn) Close() {
	self.conn.Close()
}

func NewTcpSessionConn(conn net.Conn) *TcpSessionConn {
	sc := &TcpSessionConn{}
	sc.conn = conn
	return sc
}
