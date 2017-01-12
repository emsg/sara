package utils

import (
	"fmt"
	"net"
	"time"

	"github.com/golibs/uuid"
	"github.com/tidwall/gjson"
)

var (
	success   int           = 0
	heartbeat int           = 200
	stop      chan struct{} = make(chan struct{})
	addrQueue chan string   = make(chan string, 100)
	sfpacket  string        = `{"envelope":{"pwd":"123","jid":"%s@a.a","type":0,"id":"%s"},"vsn":"0.0.1"}`
)

type client struct {
	uid, state string
	conn       net.Conn
	heartBeat  int
}

func (self *client) start() {
	packet := fmt.Sprintf(sfpacket, self.uid, uuid.Rand().Hex())
	data := append([]byte(packet), byte(1))
	if _, e := self.conn.Write(data); e == nil {
		buf := make([]byte, 256)
		i, err := self.conn.Read(buf)
		if err != nil {
			fmt.Println(i, "read_err:", err)
			return
		}
		b := buf[0 : i-1]
		r := gjson.Get(string(b), "entity.result")
		self.state = r.String()
		if "ok" != self.state {
			fmt.Println("fail", string(b))
			self.conn.Close()
		} else {
			// 保持会话，开始心跳
			go func(conn net.Conn) {
				for {
					select {
					case <-time.After(time.Second * time.Duration(self.heartBeat)):
						buf := make([]byte, 16)
						conn.Write([]byte{2, 1})
						conn.Read(buf)
					}
				}
			}(self.conn)
		}
	}
}

func newClient(addr string) (*client, error) {
	if conn, err := net.DialTimeout("tcp", addr, 3*time.Second); err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		c := &client{
			uid:       uuid.Rand().Hex(),
			conn:      conn,
			heartBeat: heartbeat,
		}
		c.start()
		return c, nil
	}
}

func genTcp() {
	for addr := range addrQueue {
		if _, e := newClient(addr); e == nil {
			success += 1
		}
	}
}

// test conn
func MakeConn(addr string, total int) {
	go genTcp()
	for i := 0; i < total; i++ {
		addrQueue <- addr
	}
	for {
		select {
		case <-time.After(time.Second * time.Duration(5)):
			fmt.Println(time.Now(), "total:", total, "success:", success, "❤️", heartbeat)
		case <-stop:
			return
		}
	}
}
