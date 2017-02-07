package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sara/config"
	"sara/core/types"
	"sara/external"
	"sara/saradb"
	"sara/utils"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/log4go"
	"github.com/golibs/uuid"
	"github.com/gorilla/websocket"
)

type ReadPacketResult struct {
	packets [][]byte
	err     error
}

func (self *ReadPacketResult) Packets() [][]byte {
	return self.packets
}
func (self *ReadPacketResult) Err() error {
	return self.err
}

type PacketHandler func(*ReadPacketResult)

type SessionConn interface {
	SetReadTimeout(timeout time.Time)
	// return packets,part,error
	ReadPacket(handler PacketHandler)
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
	SESSION_TIMEOUT        = 60
	OFFLINE_EXPIRED        = 3600 * 24 * 7 //default 7days
	SERVER_ACK      string = "server_ack"
)

type SessionStatus struct {
	Sid     string
	Jid     string
	Status  string
	Nodeid  string
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
	sync.RWMutex
	wg      *sync.WaitGroup
	Status  *SessionStatus
	sc      SessionConn
	packets chan []byte
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
	//callback_service.auth
	log4go.Debug("verify_auth jid=%s , pwd=%s", jid, pwd)
	if !external.Auth(jid_obj.GetUser(), pwd) {
		self.answer(types.NewPacketAuthFail(envelope.Id, types.FAIL_TOKEN))
		self.CloseSession("verify_check_auth")
		return
	}
	ok = true
	return
}

func (self *Session) heartbeat() {
	if self.Status.Status == types.STATUS_LOGIN {
		hb := []byte{types.HEART_BEAT}
		self.SendMessage(hb)
		//j, _ := types.NewJID(self.Status.Jid)
		//key := j.ToSessionid()
		//if _, err := self.ssdb.ResetExpire(key, SESSION_TIMEOUT); err != nil {
		//	self.CloseSession("heart_beat_on_lost_session")
		//}
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

func (self *Session) RoutePacketList(packetList []*types.Packet) {
	for _, packet := range packetList {
		self.RoutePacket(packet)
	}
}

func (self *Session) RoutePacket(packet *types.Packet) {
	self.storePacket(packet)
	jid, _ := types.NewJID(self.Status.Jid)
	// packet 里面的 from 一定是正确的,这是 SDK 决定的
	id, from, to, _ := packet.EnvelopeIdFromToType()
	switch {
	case jid.EqualWithoutResource(to):
		//给我的消息, ack 消息
		self.SendMessage(packet.ToJson())
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
				//offline line message
				log4go.Debug("📮  %s", to_key)
				external.OfflineCallback(string(packet.ToJson()))
			}
		}
	default:
		log4go.Error("☠️  error_match: jid=%s ; from=%s ; to=%s", jid.StringWithoutResource(), from, to)
	}
}

func (self *Session) SendMessage(data []byte) (int, error) {
	data = append(data, types.END_FLAG)
	return self.sc.WritePacket(data)
}

func (self *Session) CloseSession(tracemsg string) {
	log4go.Info("session_close at %s ; status=%s ; sid=%s ; jid=%s", tracemsg, self.Status.Status, self.Status.Sid, self.Status.Jid)
	self.clean <- self.Status.Sid
	if self.Status.Status == types.STATUS_LOGIN {
		j, _ := types.NewJID(self.Status.Jid)
		k := j.ToSessionid()
		idx := []byte(self.Status.Nodeid)
		self.ssdb.DeleteByIdxKey(idx, k)
	}
	self.Status.Status = types.STATUS_CLOSE
	self.sc.Close()
}

// 这个方法只能处理 c2s 的请求，并不能处理 s2s
//TODO 将 packet channel 变成普通数组传递进来，可取消一条线程,用 sc 来回调此函数
func (self *Session) packetHandler(result *ReadPacketResult) {
	defer self.setSessionTimeout()
	if result.Err() != nil {
		if self.Status.Status != types.STATUS_CLOSE {
			self.CloseSession(fmt.Sprintf("packet_handler :: %s", result.Err()))
		}
		return
	}
	for _, p := range result.Packets() {
		log4go.Debug("🚩 🚩  packets_handler => %s", p)
		if self.Status.Status == types.STATUS_CONN {
			//登陆
			self.openSession(p)
		} else if len(p) == 1 && p[0] == types.HEART_BEAT {
			//心跳
			self.heartbeat()
		} else if packet, err := types.NewPacket(p); err == nil {
			//消息协议解析,再分别处理 server_ack 和
			packet.Envelope.Ct = fmt.Sprintf("%d", utils.Timestamp13())
			id, from, to, msgtype := packet.EnvelopeIdFromToType()
			log4go.Debug("recv: %s->%s", from, p)
			if msgtype == types.MSG_TYPE_STATE && SERVER_ACK == to {
				//server_ack 消息，删除离线
				self.serverAck(packet)
			} else if msgtype == types.MSG_TYPE_CHAT {
				// 单聊
				self.answer(types.NewPacketAck(id))
				self.RoutePacket(packet)
			} else if msgtype == types.MSG_TYPE_GROUP_CHAT {
				// 群聊
				if packets, gerr := GenerateGroupPackets(self.ssdb, packet); gerr == nil {
					self.answer(types.NewPacketAck(id))
					self.RoutePacketList(packets)
				} else {
					self.answer(types.NewPacketSysNotify(id, gerr.Error()))
				}
			} else {
				//TODO 错误的操作
			}
		} else {
			log4go.Debug("👀  s=>>  %s", p)
			self.answer(types.NewPacketSysNotify(uuid.Rand().Hex(), err.Error()))
			self.CloseSession("receive_message")
		}
	}
}

/*
func (self *Session) receive() {
	self.wg.Add(1)
	defer func() {
		log4go.Info("session_done")
		self.wg.Done()
		if err := recover(); err != nil {
			log4go.Error("err ==> %v", err)
		}
	}()
	log4go.Info("recevice_started")
	self.setSessionTimeout()
	for {
		log4go.Debug("loop")
		select {
		case p := <-self.packets:
			log4go.Debug("🚩 🚩  packets_handler => %s", p)
			if self.Status.Status == types.STATUS_CONN {
				//登陆
				self.openSession(p)
			} else if len(p) == 1 && p[0] == types.HEART_BEAT {
				//心跳
				self.heartbeat()
			} else if packet, err := types.NewPacket(p); err == nil {
				//消息协议解析,再分别处理 server_ack 和
				packet.Envelope.Ct = fmt.Sprintf("%d", utils.Timestamp13())
				id, from, to, msgtype := packet.EnvelopeIdFromToType()
				log4go.Debug("recv: %s->%s", from, p)
				if msgtype == types.MSG_TYPE_STATE && SERVER_ACK == to {
					//server_ack 消息，删除离线
					self.serverAck(packet)
				} else if msgtype == types.MSG_TYPE_CHAT {
					// 单聊
					self.answer(types.NewPacketAck(id))
					self.RoutePacket(packet)
				} else if msgtype == types.MSG_TYPE_GROUP_CHAT {
					// 群聊
					if packets, gerr := GenerateGroupPackets(self.ssdb, packet); gerr == nil {
						self.answer(types.NewPacketAck(id))
						self.RoutePacketList(packets)
					} else {
						self.answer(types.NewPacketSysNotify(id, gerr.Error()))
					}
				} else {
					//TODO 错误的操作
				}
			} else {
				log4go.Debug("👀  s=>>  %s", p)
				self.answer(types.NewPacketSysNotify(uuid.Rand().Hex(), err.Error()))
				self.CloseSession("receive_message")
			}
		case result := <-self.sc.ReadPacket():
			log4go.Debug("🚩  1 session_recv_packet => %s", result)
			if result.Err() != nil {
				log4go.Debug("🚩  session_recv_packet_err")
				if self.Status.Status != types.STATUS_CLOSE {
					self.CloseSession(fmt.Sprintf("conn_error : %s", result.Err()))
				}
				return
			} else {
				log4go.Debug("🚩 2 session_recv_packet_normal")
				for _, packet := range result.Packets() {
					log4go.Debug("🚩 3 session_recv_packet=> %s", packet)
					self.packets <- packet
				}
			}
		}
		self.setSessionTimeout()
	}
}
*/

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
	idx := []byte(self.Status.Nodeid)
	// XXX : session 不应该在数据库中超时，heartbeat 只是为了刷新 tcp conn
	//self.ssdb.PutExWithIdx(idx, key, val, SESSION_TIMEOUT)
	self.ssdb.PutExWithIdx(idx, key, val, -1)
}

//所有消息都先存储起来
func (self *Session) storePacket(packet *types.Packet) {
	StorePacket(self.ssdb, packet)
}

func (self *Session) fetchOfflinePacket() (pks []*types.BasePacket, ids []string, err error) {
	jid, _ := types.NewJID(self.Status.Jid)
	idx := jid.ToOfflineKey()
	var ddata [][]byte
	if ddata, err = self.ssdb.GetByIdx(idx); err == nil {
		for _, data := range ddata {
			log4go.Debug("%s", data)
			if pk, pk_err := types.NewBasePacket(data); pk_err == nil {
				ids = append(ids, pk.Envelope.Id)
				pks = append(pks, pk)
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
	session := newSession(c, ssdb, node, cleanSession, wg)
	sc := NewTcpSessionConn(conn)
	sc.ReadPacket(session.packetHandler)
	session.sc = sc
	session.setSessionTimeout()
	return session
}

//通过 websocket 创建 session
func NewWsSession(c string, conn *websocket.Conn, ssdb saradb.Database, node MessageRouter, cleanSession chan<- string, wg *sync.WaitGroup) *Session {
	session := newSession(c, ssdb, node, cleanSession, wg)
	sc := NewWsSessionConn(conn)
	sc.ReadPacket(session.packetHandler)
	session.sc = sc
	session.setSessionTimeout()
	return session
}

func newSession(c string, ssdb saradb.Database, node MessageRouter, cleanSession chan<- string, wg *sync.WaitGroup) *Session {
	sid := uuid.Rand().Hex()
	nodeid := config.GetString("nodeid", "")
	session := &Session{
		wg:     wg,
		Status: &SessionStatus{Sid: sid, Status: types.STATUS_CONN, Nodeid: nodeid, Channel: c},
		clean:  cleanSession,
		ssdb:   ssdb,
		node:   node,
		//sc:      sc,
		packets: make(chan []byte, 32),
	}
	//go session.receive()
	return session
}

func NewSessionStatusFromJson(data []byte) *SessionStatus {
	ss := &SessionStatus{}
	if err := json.Unmarshal(data, ss); err != nil {
		log4go.Error(err)
	}
	return ss
}

func genGroupPackets(users []byte, packet *types.Packet) (packets []*types.Packet) {
	from := packet.Envelope.From
	fromJid, _ := types.NewJID(from)
	fromUser := fromJid.GetUser()
	domain := fromJid.GetDomain()
	for _, toUser := range strings.Split(string(users), ",") {
		if fromUser != toUser {
			to_jid, _ := types.NewJIDByUidDomain(toUser, domain)
			to := to_jid.StringWithoutResource()
			envelope := types.Envelope{
				Id:   uuid.Rand().Hex(),
				Type: types.MSG_TYPE_GROUP_CHAT,
				From: from,
				To:   to,
				Gid:  packet.Envelope.Gid,
				Ct:   fmt.Sprintf("%d", utils.Timestamp13()),
			}
			new_packet := &types.Packet{}
			new_packet.Envelope = envelope
			new_packet.Payload = packet.Payload
			new_packet.Vsn = packet.Vsn
			packets = append(packets, new_packet)
		}
	}
	return
}

func StorePacket(ssdb saradb.Database, packet *types.Packet) {
	id, _, to, _ := packet.EnvelopeIdFromToType()
	to_jid, _ := types.NewJID(to)
	idx := to_jid.ToOfflineKey()
	val := packet.ToJson()
	ssdb.PutExWithIdx(idx, []byte(id), val, OFFLINE_EXPIRED)
}
func GenerateGroupPackets(ssdb saradb.Database, packet *types.Packet) ([]*types.Packet, error) {
	gid := packet.Envelope.Gid
	from := packet.Envelope.From
	from_jid, _ := types.NewJID(from)
	groupUsersKey := fmt.Sprintf("group_%s@%s", gid, from_jid.GetDomain())
	fn := func(users []byte, packet *types.Packet) ([]*types.Packet, error) {
		uid := from_jid.GetUser()
		if bytes.Contains(users, []byte(uid)) {
			packets := genGroupPackets(users, packet)
			return packets, nil
		} else {
			return nil, errors.New("no_permission")
		}
	}
	if users, err := ssdb.Get([]byte(groupUsersKey)); err == nil {
		// 在缓存里寻找群成员
		log4go.Debug("gid=%s ; users=%s", gid, users)
		return fn(users, packet)
	} else {
		//callback
		ulist, ulist_err := external.GetGroupUserList(gid)
		if ulist_err == nil && len(ulist) > 1 { //群里面至少得有2个人
			us := strings.Join(ulist, ",")
			// 缓存群成员1小时
			ssdb.PutEx([]byte(groupUsersKey), []byte(us), 3600)
			return fn([]byte(us), packet)
		} else {
			log4go.Error(ulist_err)
		}
	}
	return nil, errors.New("group_notfound")
}
