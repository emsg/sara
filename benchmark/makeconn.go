package benchmark

import (
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"sara/core"
	"sara/core/types"
	"sara/utils"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alecthomas/log4go"
	"github.com/golibs/uuid"
	"github.com/tidwall/gjson"
)

var (
	wg              *sync.WaitGroup = new(sync.WaitGroup)
	fail            int             = 0
	heartbeat       int             = 50 // default gap : 50s
	messageGap      int             = 5  // default gap : 5s
	messageSize     int             = 0  // default unit k , 1 == 1024 byte
	content         []byte               // make([]byte,messageSize * 1024)
	stop            chan struct{}   = make(chan struct{})
	addrQueue       chan string     = make(chan string, 100)
	finish          chan int64      = make(chan int64)
	localAddr       string
	sfpacket        string = `{"envelope":{"pwd":"123","jid":"%s@a.a","type":0,"id":"%s"},"vsn":"0.0.1"}`
	messageTemplate string = `{"envelope":{"from":"%s@a.a","to":"%s@a.a","type":1,"id":"%s"},"vsn":"0.0.1","payload":{"content":"%s"}}`
	messageAck      string = `{"envelope":{"type":3,"from":"%s","to":"server_ack","id":"%s"},"vsn":"0.0.1"}`
	localPortPoolCh chan int
	si              *sessionIndex
	sendMessage     bool = false
	sendMessageTs   int64
)

type sessionIndex struct {
	uidMap          map[string]int // all jid reg here,use for send target
	uidArr          []string
	send, recv, per int32 //counter
	lock            *sync.RWMutex
	r               *rand.Rand
}

func (self *sessionIndex) add(uid string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if _, ok := self.uidMap[uid]; !ok {
		self.uidArr = append(self.uidArr, uid)
		self.uidMap[uid] = len(self.uidArr) - 1
	}
}
func (self *sessionIndex) del(uid string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if idx, ok := self.uidMap[uid]; ok {
		self.uidArr[idx] = ""
		delete(self.uidMap, uid)
	}
}

//发消息时，从这里随机一个 to
func (self *sessionIndex) rand(uid string) (to string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	l := len(self.uidArr)
	if l <= 0 {
		return
	}
	idx := self.r.Intn(int(len(self.uidArr) - 1))
	to = self.uidArr[idx]
	if uid == to {
		to = ""
	}
	return
}
func (self *sessionIndex) counter(action, from, to string) {
	switch action {
	case "W":
		atomic.AddInt32(&self.send, 1)
	case "R":
		atomic.AddInt32(&self.recv, 1)
	}
}

func (self *sessionIndex) showStatus() {
	sendMessage = true
	ts := utils.Timestamp10()
	fmt.Println("----------------------------", ts)
	fmt.Println("SEND\tRECV\tSEND/s\tRECV/s\tNOW")
	for {
		if self.recv > 0 && self.send > 0 {
			time.Sleep(time.Duration(time.Second * 2))
			//第一次不算
			now := utils.Timestamp10()
			_ts := now - ts
			_st := atomic.LoadInt32(&self.send)
			_rt := atomic.LoadInt32(&self.recv)
			fmt.Printf("\r\b%d\t%d\t%d\t%d\t%d", _st, _rt, (int64(_st) / _ts), (int64(_rt) / _ts), now)
		} else {
			time.Sleep(time.Duration(time.Second * 2))
		}
	}
}

func newSessionIndex() *sessionIndex {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	si := &sessionIndex{
		uidMap: make(map[string]int),
		uidArr: make([]string, 0),
		r:      r,
		lock:   new(sync.RWMutex),
	}
	return si
}

type client struct {
	uid, state string
	conn       net.Conn
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
			//reg jid
			si.add(self.uid)
			sc := core.NewTcpSessionConn(self.conn)
			// 保持会话，开始心跳
			_stop := make(chan int)
			go func(sc core.SessionConn, s chan int) {
				//heartbeat
			EndH:
				for {
					select {
					case <-s:
						break EndH
					case <-time.After(time.Second * time.Duration(heartbeat)):
						p := []byte{2, 1}
						sc.WritePacket(p)
					}
				}
			}(sc, _stop)
			if len(content) > 0 {
				go func(sc core.SessionConn, s chan int) {
					//write thread
					//fmt.Printf("👿  write thread started. content_len=%d", len(content))
				EndW:
					for {
						select {
						case <-s:
							break EndW
						case <-time.After(time.Second * time.Duration(messageGap)):
							if len(content) > 0 && sendMessage {
								id := uuid.Rand().Hex()
								from := self.uid
								if to := si.rand(self.uid); to != "" {
									message := fmt.Sprintf(messageTemplate, from, to, id, content)
									p := append([]byte(message), byte(1))
									sc.WritePacket(p)
									si.counter("W", from, to)
								}
							}
						}
					}
				}(sc, _stop)
			}
			sc.ReadPacket(func(r *core.ReadPacketResult) {
				if r.Err() != nil {
					si.del(self.uid)
					close(_stop)
				} else {
					for _, packet := range r.Packets() {
						if p, err := types.NewPacket(packet); err == nil {
							id, fm, to, tp := p.EnvelopeIdFromToType()
							if tp != types.MSG_TYPE_STATE {
								ack := fmt.Sprintf(messageAck, to, id)
								p := append([]byte(ack), byte(1))
								sc.WritePacket(p)
								si.counter("R", fm, to)
							}
						}
					}
				}
			})
		}
	}
}

func newClient(addr, laddr string, lport int) (*client, error) {
	var conn net.Conn
	var err error
	if laddr == "" {
		conn, err = net.DialTimeout("tcp", addr, 3*time.Second)
	} else {
		laddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", localAddr, lport))
		raddr, _ := net.ResolveTCPAddr("tcp", addr)
		conn, err = net.DialTCP("tcp", laddr, raddr)
	}
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		c := &client{
			uid:  uuid.Rand().Hex(),
			conn: conn,
		}
		c.start()
		return c, nil
	}
}

func genTcp() {
	for addr := range addrQueue {
		if localAddr == "" {
			if _, e := newClient(addr, "", 0); e != nil {
				fail += 1
			}
		} else if laddrs := strings.Split(localAddr, ","); len(laddrs) > 0 {
			lport := <-localPortPoolCh
			for _, laddr := range laddrs {
				if _, e := newClient(addr, laddr, lport); e != nil {
					fail += 1
				}
			}
		}
		wg.Done()
	}
}

// test conn
func MakeConn(laddr, addr string, total, hb, mg, ms int) {
	log4go.AddFilter("file", log4go.Level(log4go.ERROR), log4go.NewFileLogWriter("/tmp/sara_benchmark.log", false))
	log4go.AddFilter("stdout", log4go.Level(log4go.ERROR), log4go.NewConsoleLogWriter())
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	si = newSessionIndex()
	fmt.Println(laddr, addr, total, hb)
	localAddr = laddr
	if laddr != "" {
		localPortPoolCh = make(chan int, 65535)
		for i := 65535; i > 65535-total-100; i-- {
			localPortPoolCh <- i
		}
	}
	heartbeat, messageGap, messageSize = hb, mg, ms
	//init message
	if messageSize > 0 {
		content = make([]byte, messageSize*1024)
		for i, _ := range content {
			content[i] = byte(97)
		}
	}
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
	ll := len(strings.Split(localAddr, ","))
	fmt.Println("cpu core:", cpu, " worker:", cpu*2)
	fmt.Println("😊  total:", total*ll, "finished , fail:", fail, " time:", (e - s), "ms. heartbeat:", hb)

	if messageSize > 0 {
		go si.showStatus()
	}
	close(addrQueue)
	close(finish)
	<-stop
}
