package node

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"sara/config"
	"sara/core"
	"sara/core/types"
	"sara/saradb"
	"sara/sararpc"
	"sara/utils"
	"sync"

	"github.com/alecthomas/log4go"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
)

type WSHandler struct {
	node *Node
}

func (self *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.node.acceptWs(w, r)
}

type Node struct {
	lock                           *sync.RWMutex
	wg                             *sync.WaitGroup
	Nodeid, name                   string
	sessionMap                     map[string]*core.Session //all avaliable session
	Port, TLSPort, WSPort, WSSPort int
	stop                           chan int
	cleanSession                   chan string
	tcpListen                      *net.TCPListener
	tlsListen                      net.Listener
	wsListen, wssListen            net.Listener
	db                             saradb.Database //SessionStatus db
	dataChannel                    sararpc.DataChannel
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (self *Node) GetDB() saradb.Database {
	return self.db
}

func (self *Node) StartWS() error {
	listen, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("0.0.0.0"), self.WSPort, ""})
	if err != nil {
		log4go.Error("Fail start ws; err = %s", err)
		return err
	}
	self.wsListen = listen
	//addr := fmt.Sprintf("0.0.0.0:%d", self.WSPort)
	go func() {
		//http.HandleFunc("/", self.acceptWs)
		http.Serve(listen, &WSHandler{node: self})
		//http.ListenAndServe(addr, &WSHandler{node: self})
	}()
	log4go.Info("ws start on [%d]", self.WSPort)
	return nil
}
func (self *Node) StartTCP() error {
	listen, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("0.0.0.0"), self.Port, ""})
	if err != nil {
		log4go.Error("Fail start node ; err = %s", err)
		return err
	}
	log4go.Info("tcp start on [0.0.0.0:%d]", self.Port)
	self.tcpListen = listen
	go self.acceptTCP()
	return nil
}

/*
func (self *Node) StartWSS() error {
	certfile := config.GetString("certfile", "/etc/sara/server.pem")
	keyfile := config.GetString("keyfile", "/etc/sara/server.key")
	addr := fmt.Sprintf("0.0.0.0:%d", self.WSSPort)
	go func() {
		//http.HandleFunc("/", self.acceptWs)
		http.ListenAndServeTLS(addr, certfile, keyfile, &WSHandler{node: self})
	}()
	log4go.Info("wss start on [%s]", addr)
	return nil
}
*/
func (self *Node) StartWSS() error {
	certfile := config.GetString("certfile", "/etc/sara/server.pem")
	keyfile := config.GetString("keyfile", "/etc/sara/server.key")
	cert, err := tls.LoadX509KeyPair(certfile, keyfile)
	if err != nil {
		log4go.Error("Fail start wss ; err = %s", err)
		return err
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	addr := fmt.Sprintf(":%d", self.WSSPort)
	listen, err := tls.Listen("tcp", addr, config)
	if err != nil {
		log4go.Error("Fail start wss; err = %s", err)
		return err
	}
	log4go.Info("wss start on [0.0.0.0:%d]", self.WSSPort)
	self.wssListen = listen
	go func() {
		http.Serve(listen, &WSHandler{node: self})
	}()
	log4go.Info("wss start on [%d]", self.WSSPort)
	return nil
}
func (self *Node) StartTLS() error {
	certfile := config.GetString("certfile", "/etc/sara/server.pem")
	keyfile := config.GetString("keyfile", "/etc/sara/server.key")
	cert, err := tls.LoadX509KeyPair(certfile, keyfile)
	if err != nil {
		log4go.Error("Fail start node ; err = %s", err)
		return err
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	addr := fmt.Sprintf(":%d", self.TLSPort)
	listen, err := tls.Listen("tcp", addr, config)
	if err != nil {
		log4go.Error("Fail start node ; err = %s", err)
		return err
	}
	log4go.Info("tls start on [0.0.0.0:%d]", self.TLSPort)
	self.tlsListen = listen
	go self.acceptTLS()
	return nil
}

func (self *Node) closeListener() {
	if self.tcpListen != nil {
		log4go.Warn("close tcp listen.")
		self.tcpListen.Close()
	}
	if self.tlsListen != nil {
		log4go.Warn("close tls listen.")
		self.tlsListen.Close()
	}
	if self.wsListen != nil {
		log4go.Warn("close ws listen.")
		self.wsListen.Close()
	}
	if self.wssListen != nil {
		log4go.Warn("close wss listen.")
		self.wssListen.Close()
	}
}

func (self *Node) Wait() {
	//if self.tcpListen == nil {
	//	return
	//}
	<-self.stop
	self.closeListener() // 关闭所有服务端口，停止接收新session
	log4go.Info("⛑️  begin security shutdown.")
	//shutdown
	i := 0
	for _, s := range self.sessionMap {
		s.CloseSession("node_stop")
		i++
	}
	log4go.Info("please wait session close. total=%d , clean=%d", len(self.sessionMap), i)
	self.wg.Wait()
	log4go.Info("session close success")
	self.db.Close()
	//defer self.tcpListen.Close()
	log4go.Info("security shutdown success.")
}

func (self *Node) Stop() {
	self.stop <- 1
}

func (self *Node) clean() {
	defer func() {
		recover()
	}()
	for sid := range self.cleanSession {
		self.lock.Lock()
		delete(self.sessionMap, sid)
		log4go.Debug("clean_session_success sid=%s", sid)
		self.lock.Unlock()
	}
}
func (self *Node) fetchSession(sid string) (session *core.Session, ok bool) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	session, ok = self.sessionMap[sid]
	return
}
func (self *Node) registerSession(session *core.Session) {
	if sid := session.Status.Sid; sid != "" {
		log4go.Debug("reg_session sid=%s", sid)
		self.lock.Lock()
		self.sessionMap[sid] = session
		self.lock.Unlock()
	}
}

//implements MessageRouter interface >>>>>>>>
func (self *Node) Route(channel, sid string, packet *types.Packet, signal ...byte) {
	if channel == "" || self.dataChannel.GetChannel() == channel {
		if session, ok := self.fetchSession(sid); ok {
			if signal != nil {
				log4go.Debug("👮 node.route_signal -> %s", signal)
				if signal[0] == types.KILL {
					session.Kill()
				}
			} else {
				log4go.Debug("👮 node.route_packet-> %s", packet.ToJson())
				session.RoutePacket(packet)
			}
		}
	} else {
		self.dataChannel.Publish(channel, string(packet.ToJson()))
	}
}

func (self *Node) IsCurrentChannel(n string) bool {
	if n == self.dataChannel.GetChannel() {
		return true
	} else {
		return false
	}
}

//implements MessageRouter interface <<<<<<<

//接收来自 c 端的请求
func (self *Node) acceptTCP() {
	c := self.dataChannel.GetChannel()
	for {
		// 阻塞在这里，
		//if conn, err := self.tcpListen.Accept(); err == nil {
		if conn, err := self.tcpListen.AcceptTCP(); err == nil {
			//conn.SetKeepAlive(true)
			self.registerSession(core.NewTcpSession(c, conn, self.db, self, self.cleanSession, self.wg))
		}
	}
}

func (self *Node) acceptTLS() {
	c := self.dataChannel.GetChannel()
	for {
		// 阻塞在这里，
		//if conn, err := self.tcpListen.Accept(); err == nil {
		if conn, err := self.tlsListen.Accept(); err == nil {
			//conn.SetKeepAlive(true)
			self.registerSession(core.NewTcpSession(c, conn, self.db, self, self.cleanSession, self.wg))
		}
	}
}
func (self *Node) acceptWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log4go.Debug("🌍  >>> err_upgrade: %s", err)
		return
	}
	log4go.Debug("🌍  >>> upgrade success")
	c := self.dataChannel.GetChannel()
	self.registerSession(core.NewWsSession(c, conn, self.db, self, self.cleanSession, self.wg))
}

func (self *Node) dataChannelHandler(message string) {
	log4go.Debug("☕️ ----> %s", message)
	packet, err := types.NewPacket([]byte(message))
	if err != nil {
		log4go.Error("error packet : %s", message)
		return
	}
	//TODO type=2
	switch packet.Envelope.Type {
	case types.MSG_TYPE_GROUP_CHAT:
		if packets, gerr := core.GenerateGroupPackets(self.db, packet); gerr == nil {
			for _, p := range packets {
				log4go.Debug("%s", p.ToJson())
				self.routePacket(p)
			}
		} else {
			log4go.Error(gerr.Error())
		}
	default:
		//type!=2
		self.routePacket(packet)
	}
}

func (self *Node) routePacket(packet *types.Packet) {
	if jid, jid_err := types.NewJID(packet.Envelope.To); jid_err == nil {
		skey := jid.ToSessionid()
		if ssb, se := self.db.Get(skey); se == nil {
			ss := core.NewSessionStatusFromJson(ssb)
			self.Route(ss.Channel, ss.Sid, packet)
		} else {
			core.StorePacket(self.db, packet)
			log4go.Debug("☕️  📮  %s", skey)
		}
	}
}
func (self *Node) cleanGhostSession() {
	//XXX clean ghost session
	nodeid := []byte(self.Nodeid)
	ts := fmt.Sprintf("%d", utils.Timestamp13())
	//所有的节点，启动后，都注册在 sara 这个 hashtable 中,记录节点的启动时间
	self.db.PutExWithIdx([]byte("sara"), nodeid, []byte(ts), 0)
	log4go.Info("register node : %s", nodeid)
	self.db.DeleteByIdx(nodeid)
	log4go.Info("🔪  👻  clean ghost session")
}

//TODO 如何控制 nodeid 全局唯一，是否需要 gossip ?
func New(ctx *cli.Context) *Node {
	node := &Node{
		lock:         &sync.RWMutex{},
		wg:           &sync.WaitGroup{},
		sessionMap:   make(map[string]*core.Session),
		cleanSession: make(chan string, 4096),
		Port:         config.GetInt("port", 4222),
		WSPort:       config.GetInt("wsport", 4224),
		TLSPort:      config.GetInt("tlsport", 4333),
		WSSPort:      config.GetInt("wssport", 4334),
		stop:         make(chan int),
	}
	dbaddr := config.GetString("dbaddr", "localhost:6379")
	dbpool := config.GetInt("dbpool", 100)

	rpcserverAddr := config.GetString("nodeaddr", "localhost:4281")
	node.name = rpcserverAddr

	if db, err := saradb.NewClusterDatabase(dbaddr, dbpool); err != nil {
		if db, err = saradb.NewDatabase(dbaddr, dbpool); err != nil {
			log4go.Error(err)
			return nil
		} else {
			node.db = db
		}
	} else {
		node.db = db
	}
	//node.dataChannel = node.db.GenDataChannel(node.name)
	//node.dataChannel.Subscribe(node.dataChannelHandler)
	node.dataChannel = sararpc.NewRPCDataChannel(rpcserverAddr, 20000)
	node.dataChannel.Subscribe(node.dataChannelHandler)
	if rpcserver, err := sararpc.NewRPCServer(rpcserverAddr, node.dataChannel); err != nil {
		defer os.Exit(0)
		log4go.Error("rpcserver start fail . err=%v", err)
	} else {
		go func() { rpcserver.Start() }()
	}
	node.Nodeid = config.GetString("nodeid", "")
	node.cleanGhostSession()
	go node.clean()
	return node
}
