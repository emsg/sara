package core

import (
	"errors"
	"github.com/alecthomas/log4go"
	"net"
	"time"
)

var max_buf_len int = 4096

type TcpSessionConn struct {
	handler PacketHandler
	conn    net.Conn
	part    []byte
	//resultCh chan *ReadPacketResult
}

func (self *TcpSessionConn) SetReadTimeout(timeout time.Time) {
	self.conn.SetReadDeadline(timeout)
	//self.conn.SetWriteDeadline(timeout)
}
func (self *TcpSessionConn) callbackHandler(r *ReadPacketResult) {
	defer func() {
		if e := recover(); e != nil {
			log4go.Error("❌  null point exception ::> %s ::> %s", e, r)
			self.conn.Close()
		}
	}()
	self.handler(r)
}

func (self *TcpSessionConn) recv() {
	defer func() {
		if e := recover(); e != nil {
			log4go.Error("☠️  %s", e)
			switch e.(type) {
			case error:
				self.callbackHandler(&ReadPacketResult{
					err: e.(error),
				})
			}
		}
	}()
	var buff_len int = 1024
	for {
		//动态调整缓冲区
		buff := make([]byte, buff_len)
		n, e := self.conn.Read(buff)
		MeasureReadAdd(1)
		if e != nil {
			self.callbackHandler(&ReadPacketResult{
				err: e,
			})
			return
		}
		log4go.Debug("buff_len=%d,n=%d", buff_len, n)
		packetList, newPart, err := DecodePacket(buff[:n], self.part)
		self.part = newPart
		if len(packetList) > 0 {
			self.callbackHandler(&ReadPacketResult{
				packets: packetList,
				err:     err,
			})
		}
		change_buff_len(&buff_len, n)
	}
}

func (self *TcpSessionConn) ReadPacket(handler PacketHandler) {
	self.handler = handler
	go self.recv()
}

func (self *TcpSessionConn) WritePacket(packet []byte) (i int, e error) {
	defer func() {
		if err := recover(); err != nil {
			log4go.Debug(err)
			i, e = 0, errors.New(err.(string))
		}
	}()
	i, e = self.conn.Write(packet)
	MeasureWriteAdd(1)
	return
}

func (self *TcpSessionConn) Close() {
	self.conn.Close()
	//close(self.resultCh)
}

func NewTcpSessionConn(conn net.Conn) *TcpSessionConn {
	sc := &TcpSessionConn{
		part: make([]byte, 0),
		//resultCh: make(chan *ReadPacketResult),
		conn: conn,
	}
	return sc
}
func change_buff_len(buff_len *int, n int) {
	if n > 2 {
		// n == 2 时，可能是心跳和 kill 之类的信号，并非 packet
		switch {
		case n == *buff_len:
			//放大
			*buff_len += (*buff_len / 4)
		case n < *buff_len:
			//缩小
			*buff_len -= (*buff_len / 4)
		}
		switch {
		case *buff_len > max_buf_len:
			*buff_len = max_buf_len
		case *buff_len < n:
			*buff_len = n + (n / 4)
		}
	}
}
