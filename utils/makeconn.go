package utils

import (
	"fmt"
	"github.com/golibs/uuid"
	"github.com/tidwall/gjson"
	"net"
	"runtime"
	"sync"
	"time"
)

var (
	wg              *sync.WaitGroup = new(sync.WaitGroup)
	fail            int             = 0
	heartbeat       int             = 50
	stop            chan struct{}   = make(chan struct{})
	addrQueue       chan string     = make(chan string, 100)
	finish          chan int64      = make(chan int64)
	localAddr       string
	sfpacket        string = `{"envelope":{"pwd":"123","jid":"%s@a.a","type":0,"id":"%s"},"vsn":"0.0.1"}`
	localPortPoolCh chan int
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
	var conn net.Conn
	var err error
	if localAddr == "" {
		conn, err = net.DialTimeout("tcp", addr, 3*time.Second)
	} else {
		lport := <-localPortPoolCh
		laddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", localAddr, lport))
		raddr, _ := net.ResolveTCPAddr("tcp", addr)
		conn, err = net.DialTCP("tcp", laddr, raddr)
	}
	if err != nil {
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
		if _, e := newClient(addr); e != nil {
			fail += 1
		}
		wg.Done()
	}
}

// test conn
func MakeConn(laddr, addr string, total, hb int) {
	fmt.Println(laddr, addr, total, hb)
	localAddr = laddr
	if laddr != "" {
		localPortPoolCh = make(chan int, 65535)
		for i := 65535; i > 65535-total-100; i-- {
			localPortPoolCh <- i
		}
	}
	heartbeat = hb
	cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(cpu)
	for i := 0; i < cpu*2; i++ {
		go genTcp()
	}
	s := time.Now().UnixNano() / 1000000
	wg.Add(total)
	for i := 0; i < total; i++ {
		addrQueue <- addr
	}
	wg.Wait()
	e := time.Now().UnixNano() / 1000000
	fmt.Println("cpu core:", cpu, " worker:", cpu*2)
	fmt.Println("😊  total:", total, "finished , fail:", fail, " time:", (e - s), "ms. heartbeat:", hb)
	close(addrQueue)
	close(finish)
	<-stop
}
