package node

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sara/core"
	"sara/core/types"
	"sara/saradb"
	"sara/sararpc"
	"strconv"

	"github.com/alecthomas/log4go"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
)

type Node struct {
	name                  string
	sessionMap            map[string]*core.Session //all avaliable session
	Port, SSLPort, WSPort int
	stop                  chan struct{}
	cleanSession          chan string
	tcpListen             net.Listener
	db                    saradb.Database //SessionStatus db
	dataChannel           sararpc.DataChannel
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (self *Node) StartWS() error {
	addr := fmt.Sprintf("0.0.0.0:%d", self.WSPort)
	go func() {
		http.HandleFunc("/", self.acceptWs)
		http.ListenAndServe(addr, nil)
	}()
	log4go.Info("ws start on [%s]", addr)
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

func (self *Node) Wait() {
	defer self.tcpListen.Close()
	stop := self.stop
	if self.tcpListen == nil {
		return
	}
	<-stop
}

func (self *Node) clean() {
	for sid := range self.cleanSession {
		log4go.Debug("clean_session sid=%s", sid)
		delete(self.sessionMap, sid)
	}
}
func (self *Node) registerSession(session *core.Session) {
	if sid := session.Status.Sid; sid != "" {
		log4go.Debug("reg_session sid=%s", sid)
		self.sessionMap[sid] = session
	}
}

//implements MessageRouter interface >>>>>>>>
func (self *Node) Route(channel, sid string, packet *types.Packet, signal ...byte) {
	if channel == "" || self.dataChannel.GetChannel() == channel {
		if session, ok := self.sessionMap[sid]; ok {
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
		if conn, err := self.tcpListen.Accept(); err == nil {
			self.registerSession(core.NewTcpSession(c, conn, self.db, self, self.cleanSession))
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
	self.registerSession(core.NewWsSession(c, conn, self.db, self, self.cleanSession))
}

func (self *Node) dataChannelHandler(message string) {
	log4go.Debug("☕️ ----> %s", message)
	packet, err := types.NewPacket([]byte(message))
	if err != nil {
		log4go.Error("error packet : %s", message)
		return
	}
	if jid, jid_err := types.NewJID(packet.Envelope.To); jid_err == nil {
		skey := jid.ToSessionid()
		if ssb, se := self.db.Get(skey); se == nil {
			ss := core.NewSessionStatusFromJson(ssb)
			self.Route(ss.Channel, ss.Sid, packet)
		} else {
			log4go.Debug("☕️  📮  %s", skey)
		}
	}
}

func New(ctx *cli.Context) *Node {
	node := &Node{
		sessionMap:   make(map[string]*core.Session),
		cleanSession: make(chan string, 4096),
		Port:         ctx.GlobalInt("port"),
		WSPort:       ctx.GlobalInt("wsport"),
	}
	dbaddr := ctx.GlobalString("dbaddr")
	dbpool := ctx.GlobalInt("dbpool")

	hostname := ctx.GlobalString("hostname")
	if hostname == "" {
		defer os.Exit(0)
		log4go.Error("❌❌❌  hostname or ip can not empty, use --hostname to set.")
	}
	rpcport := ctx.GlobalInt("rpcport")
	rpcserverAddr := net.JoinHostPort(hostname, strconv.Itoa(rpcport))
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
	go node.clean()
	return node
}
