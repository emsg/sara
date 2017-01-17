package core

import (
	"encoding/json"
	"fmt"
	"net"
	"sara/core/types"
	"sara/saradb"
	"sara/utils"
	"sync"
	"time"

	"github.com/alecthomas/log4go"
	"github.com/golibs/uuid"
	"github.com/gorilla/websocket"
)

type SessionConn interface {
	SetReadTimeout(timeout time.Time)
	// return packets,part,error
	ReadPacket(part []byte) ([][]byte, []byte, error)
	WritePacket(packet []byte) (int, error)
	Close()
}

type MessageRouter interface {
	// 当目标session在线时，将packet路由到制定的node上完成发送
	Route(channel, sid string, packet *types.Packet, signal ...byte)
	// 判断传入的 node 是否等于当前节点的 endpoint
	IsCurrentChannel(node string) bool
}

const (
	LOGIN_TIMEOUT   int    = 5
	SESSION_TIMEOUT        = 300
	OFFLINE_EXPIRED        = 3600 * 24 * 7 //default 7days
	SERVER_ACK      string = "server_ack"
)

type SessionStatus struct {
	Sid     string
	Jid     string
	Status  string
	Channel string
	Ct      int64
}

func (self *SessionStatus) ToJson() []byte {
	if d, e := json.Marshal(self); e != nil {
		log4go.Error(e)
		return nil
	} else {
		return d
	}
}

type Session struct {
	wg      *sync.WaitGroup
	Status  *SessionStatus
	sc      SessionConn
	packets chan []byte
	part    []byte
	stop    chan struct{}
	clean   chan<- string
	ssdb    saradb.Database
	node    MessageRouter
}

func (self *Session) openSession(p []byte) {
	if packet, ok := self.verify(p); ok {
		envelope := packet.Envelope
		jid, pwd := envelope.Jid, envelope.Pwd
		self.Status.Jid = jid
		self.Status.Status = types.STATUS_LOGIN
		self.Status.Ct = utils.Timestamp10()
		p := types.NewPacketAuthSuccess(envelope.Id)
		p.AddDelay(nil)
		if ofl_pks, ofl_ids, ofl_err := self.fetchOfflinePacket(); ofl_err == nil {
			p.AddDelay(ofl_pks)
			//将消息id与dealy.packets 中的消息id
			//映射到内存中，等待ack时，删除这一批离线消息
			log4go.Debug(ofl_ids)
			if len(ofl_ids) > 0 {
				ofl_cache.put(envelope.Id, ofl_ids)
			}
		}
		log4go.Debug("login jid=%s , pwd=%s , %s", jid, pwd, p)
		self.storeSessionStatus()
		self.answer(p)
	}
}

// timeout规定时间内没有发送 “打开会话” 请求
// fail_type第一个请求应该是打开会话请求，即type=0，否则返回此错误
// fail_tokeninner_token过期或失效
// fail_param属性不符合规则
// fail_packet数据包与协议不符
func (self *Session) verify(p []byte) (packet *types.Packet, ok bool) {
	var err error
	packet, err = types.NewPacket(p)
	if err != nil {
		self.answer(types.NewPacketAuthFail(uuid.Rand().Hex(), types.FAIL_PACKET))
		self.CloseSession("verify_decode_packet")
		return
	}
	log4go.Debug("verify = %s", packet)
	envelope := packet.Envelope

	if envelope.Type != types.MSG_TYPE_OPEN_SESSION {
		self.answer(types.NewPacketAuthFail(envelope.Id, types.FAIL_TYPE))
		self.CloseSession("verify_check_msgtype")
		return
	}
	jid, pwd := envelope.Jid, envelope.Pwd
	// check jid
	jid_obj, jid_err := types.NewJID(jid)
	if jid_err != nil {
		self.answer(types.NewPacketAuthFail(envelope.Id, types.FAIL_PACKET))
		self.CloseSession("verify_check_jid")
		return
	}
	// check already logon
	if s, e := self.ssdb.Get(jid_obj.ToSessionid()); e == nil {
		ss := NewSessionStatusFromJson(s)
		//TODO 这里可以根据用户个人配置，来决定T掉哪一个
		//默认踢掉前一个
		self.node.Route(ss.Channel, ss.Sid, nil, types.KILL)
	}
	//TODO callback_service.auth
	log4go.Debug("verify_auth jid=%s , pwd=%s", jid, pwd)
	ok = true
	return
}

func (self *Session) heartbeat() {
	if self.Status.Status == types.STATUS_LOGIN {
		hb := []byte{types.HEART_BEAT}
		self.SendMessage(hb)
		self.storeSessionStatus()
		log4go.Debug("❤️  %s ->%d", self.Status.Jid, hb)
	}
}

func (self *Session) answer(packet *types.Packet) {
	jp := packet.ToJson()
	log4go.Debug("answer-> %s", jp)
	self.SendMessage(jp)
}

func (self *Session) Kill() {
	log4go.Info("☠️  Repeat login")
	self.SendMessage([]byte{types.KILL})
	self.CloseSession("kill")
}

func (self *Session) RoutePacket(packet *types.Packet) {
	jid, _ := types.NewJID(self.Status.Jid)
	// packet 里面的 from 一定是正确的,这是 SDK 决定的
	id, from, to, _ := packet.EnvelopeIdFromToType()
	switch {
	case jid.EqualWithoutResource(to):
		//给我的消息, ack 消息
		self.messageHandler(packet)
	case jid.EqualWithoutResource(from):
		//我发出去的消息
		if to_jid, err := types.NewJID(to); err != nil {
			notify := types.NewPacketSysNotify(id, err.Error())
			self.answer(notify)
		} else {
			to_key := to_jid.ToSessionid()
			//find target session
			if ssb, se := self.ssdb.Get(to_key); se == nil {
				ss := NewSessionStatusFromJson(ssb)
				self.node.Route(ss.Channel, ss.Sid, packet)
				log4go.Debug("✉️  %s->%s", to_key, ssb)
			} else {
				log4go.Debug("📮  %s", to_key)
			}
		}
	}
}

func (self *Session) SendMessage(data []byte) (int, error) {
	data = append(data, types.END_FLAG)
	return self.sc.WritePacket(data)
}

func (self *Session) CloseSession(tracemsg string) {
	log4go.Debug("session_close at %s ; sid=%s", tracemsg, self.Status.Sid)
	self.clean <- self.Status.Sid
	if self.Status.Status == types.STATUS_LOGIN {
		j, _ := types.NewJID(self.Status.Jid)
		k := j.ToSessionid()
		self.ssdb.Delete(k)
	}
	self.Status.Status = types.STATUS_CLOSE
	self.sc.Close()
}

//我收到的其他 session 发给我的消息
func (self *Session) messageHandler(packet *types.Packet) {
	log4go.Debug("messageHandler -> %s", packet)
	self.SendMessage(packet.ToJson())
}

// 这个方法只能处理 c2s 的请求，并不能处理 s2s
func (self *Session) receive() {
	self.wg.Add(1)
	defer func() {
		if err := recover(); err != nil {
			log4go.Debug("err ==> %v", err)
		}
		self.wg.Done()
	}()
	log4go.Debug("receive_started")
	for {
		select {
		case <-self.stop:
			log4go.Debug("session_stop")
		case p := <-self.packets:
			if self.Status.Status == types.STATUS_CONN {
				//登陆
				self.openSession(p)
			} else if len(p) == 1 && p[0] == types.HEART_BEAT {
				//心跳
				self.heartbeat()
			} else if packet, err := types.NewPacket(p); err == nil {
				//消息协议解析,再分别处理 server_ack 和
				id, from, to, msgtype := packet.EnvelopeIdFromToType()
				log4go.Debug("recv: %s->%s", from, p)
				if msgtype == types.MSG_TYPE_STATE && SERVER_ACK == to {
					//server_ack 消息，删除离线
					self.serverAck(packet)
				} else if msgtype == types.MSG_TYPE_CHAT {
					// 单聊
					self.answer(types.NewPacketAck(id))
					self.storePacket(packet)
					self.RoutePacket(packet)
				} else if msgtype == types.MSG_TYPE_GROUP_CHAT {
					//TODO 群聊
				} else {
					//TODO 错误的操作
				}
			} else {
				log4go.Debug("👀  s=>>  %s", p)
				log4go.Debug("👀  b=>>  %b", p)
				self.answer(types.NewPacketSysNotify(uuid.Rand().Hex(), err.Error()))
				self.CloseSession("receive_message")
			}
		default:
			self.setSessionTimeout()
			if packetList, part, err := self.sc.ReadPacket(self.part); err != nil {
				switch err.Error() {
				case "EOF":
					if self.Status.Status != types.STATUS_CLOSE {
						self.CloseSession("receive_eof_normal")
					}
				case "TIMEOUT":
					self.answer(types.NewPacketSysNotify(uuid.Rand().Hex(), types.FAIL_TIMEOUT))
					self.CloseSession("receive_timeout")
				default:
					self.CloseSession(fmt.Sprintf("receive_like_error: %s", err.Error()))
				}
				return
			} else {
				self.part = part
				for _, packet := range packetList {
					self.packets <- packet
				}
			}
		}
	}
}

func (self *Session) setSessionTimeout() {
	var t time.Time
	if self.Status.Status == types.STATUS_CONN {
		t = utils.TimeoutTime(LOGIN_TIMEOUT)
	} else {
		t = utils.TimeoutTime(SESSION_TIMEOUT)
	}
	self.sc.SetReadTimeout(t)
}
func (self *Session) storeSessionStatus() {
	jidStr := self.Status.Jid
	j, _ := types.NewJID(jidStr)
	key := j.ToSessionid()
	val := self.Status.ToJson()
	//log4go.Debug("storeSessionStatus-> %s", val)
	self.ssdb.PutEx(key, val, SESSION_TIMEOUT)
}

//所有消息都先存储起来
func (self *Session) storePacket(packet *types.Packet) {
	id, _, to, _ := packet.EnvelopeIdFromToType()
	to_jid, _ := types.NewJID(to)
	idx := to_jid.ToOfflineKey()
	val := packet.ToJson()
	self.ssdb.PutExWithIdx(idx, []byte(id), val, OFFLINE_EXPIRED)
}

func (self *Session) fetchOfflinePacket() (pks []*types.BasePacket, ids []string, err error) {
	jid, _ := types.NewJID(self.Status.Jid)
	idx := jid.ToOfflineKey()
	var ddata [][]byte
	if ddata, err = self.ssdb.GetByIdx(idx); err == nil {
		for i, data := range ddata {
			if pk, pk_err := types.NewBasePacket(data); pk_err == nil {
				ids = append(ids, pk.Envelope.Id)
				pks = append(pks, pk)
			} else {
				log4go.Debug("%d ----> err = %s", i, pk_err)
				log4go.Debug("%d ----> data = %s", i, data)
				log4go.Debug("%d ----> pk = %s", i, pk)
			}
		}
	}
	return
}

//收到ack消息就删除对应的消息
func (self *Session) serverAck(packet *types.Packet) {
	jid, _ := types.NewJID(self.Status.Jid)
	idx := jid.ToOfflineKey()
	if ids, err := ofl_cache.get(packet.Envelope.Id); err == nil {
		// serverack for offline message
		for _, id := range ids {
			key := []byte(id)
			self.ssdb.DeleteByIdxKey(idx, key)
		}
	} else {
		key := []byte(packet.Envelope.Id)
		self.ssdb.DeleteByIdxKey(idx, key)
	}
}

//通过 tcp 创建 session
func NewTcpSession(c string, conn net.Conn, ssdb saradb.Database, node MessageRouter, cleanSession chan<- string, wg *sync.WaitGroup) *Session {
	sc := NewTcpSessionConn(conn)
	return newSession(c, sc, ssdb, node, cleanSession, wg)
}

//通过 websocket 创建 session
func NewWsSession(c string, conn *websocket.Conn, ssdb saradb.Database, node MessageRouter, cleanSession chan<- string, wg *sync.WaitGroup) *Session {
	sc := NewWsSessionConn(conn)
	return newSession(c, sc, ssdb, node, cleanSession, wg)
}

//TODO 通过 tls 创建 session
func NewTlsSession() {

}

func newSession(c string, sc SessionConn, ssdb saradb.Database, node MessageRouter, cleanSession chan<- string, wg *sync.WaitGroup) *Session {
	sid := uuid.Rand().Hex()
	session := &Session{
		wg:      wg,
		Status:  &SessionStatus{Sid: sid, Status: types.STATUS_CONN, Channel: c},
		clean:   cleanSession,
		ssdb:    ssdb,
		node:    node,
		sc:      sc,
		packets: make(chan []byte, 1024),
	}
	go session.receive()
	return session
}

func NewSessionStatusFromJson(data []byte) *SessionStatus {
	ss := &SessionStatus{}
	if err := json.Unmarshal(data, ss); err != nil {
		log4go.Error(err)
	}
	return ss
}
