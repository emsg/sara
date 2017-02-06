package core

import (
	"errors"
	"github.com/alecthomas/log4go"
	"net"
	"time"
)

type TcpSessionConn struct {
	conn     net.Conn
	part     []byte
	resultCh chan *ReadPacketResult
}

func (self *TcpSessionConn) SetReadTimeout(timeout time.Time) {
	self.conn.SetReadDeadline(timeout)
	self.conn.SetWriteDeadline(timeout)
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
		log4go.Debug("👀  2 tcp_read_packet_buff => %s", buff)
		packetList, newPart, _ := DecodePacket(buff, self.part)
		self.part = newPart
		log4go.Debug("👀  3 tcp_read_packet_decode => %s", packetList)
		r := &ReadPacketResult{
			packets: packetList,
			err:     e,
		}
		log4go.Debug("👀  4 tcp_read_packet_return => %s", r)
		self.resultCh <- r
		if e != nil {
			log4go.Debug("recv_error")
			return
		}
	}
}

func (self *TcpSessionConn) ReadPacket() <-chan *ReadPacketResult {
	return self.resultCh
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
	close(self.resultCh)
}

func NewTcpSessionConn(conn net.Conn) *TcpSessionConn {
	sc := &TcpSessionConn{
		part:     make([]byte, 0),
		resultCh: make(chan *ReadPacketResult),
	}
	sc.conn = conn
	go sc.recv()
	return sc
}
