package core

import (
	"errors"
	"github.com/alecthomas/log4go"
	"net"
	"time"
)

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

func (self *TcpSessionConn) recv() {
	defer func() {
		if r := recover(); r != nil {
			log4go.Error(r)
		}
	}()
	for {
		log4go.Debug("👀  1 tcp_read_packet")
		buff := make([]byte, 256)
		_, e := self.conn.Read(buff)
		if e != nil {
			log4go.Debug("recv_error => %s", e)
			r := &ReadPacketResult{
				err: e,
			}
			// XXX 是否可以异步处理？比如每次一个新的线程来 handler
			self.handler(r)
			return
		}
		log4go.Debug("👀  2 tcp_read_packet_buff %d => %b", len(buff), buff)
		packetList, newPart, err := DecodePacket(buff, self.part)
		self.part = newPart
		log4go.Debug("👀  3 tcp_read_packet_decode => %s", packetList)
		if len(packetList) > 0 {
			r := &ReadPacketResult{
				packets: packetList,
				err:     err,
			}
			log4go.Debug("👀  4 tcp_read_packet_return => %s", r)
			//self.resultCh <- r
			self.handler(r)
		}
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
