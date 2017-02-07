package saradb

import (
	"bytes"
	"errors"
	"fmt"
	"sara/core/types"
	"sync/atomic"
	//"sara/utils"
	"sync"
	"time"

	"github.com/alecthomas/log4go"
	"github.com/golibs/uuid"
	"github.com/mediocregopher/radix.v2/cluster"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

var (
	IDX_SUFFIX     = []byte("_idx")
	SESSION_PERFIX = types.SESSION_PERFIX
)

type DBMODEL int

type writeBufArgs struct {
	cmd  string
	args []interface{}
	resp chan *redis.Resp
}

const (
	MODEL_CLUSTER DBMODEL = iota
	MODEL_SINGLE
)

type SaraDatabase struct {
	Addr           string
	PoolSize       int
	c_cli          *cluster.Cluster
	p_cli          *pool.Pool
	model          DBMODEL
	stop           chan struct{}
	tl             bool
	wbCh           chan writeBufArgs //异步的写操作缓冲区
	wbCh4s         chan writeBufArgs //异步的写操作缓冲区,for session
	wbSyncChList   []chan int        //同步写操作工作线程
	wbSyncChList4s []chan int        //同步写操作工作线程,for session
	wg             *sync.WaitGroup   //同步写操作工作线程
}

func (self *SaraDatabase) wbfConsumerWorker(wbSyncCh chan int, wbCh chan writeBufArgs, _wbTotal int32) {
	wbTotal := atomic.LoadInt32(&_wbTotal)
	uuid := uuid.Rand().Hex()
	defer self.wg.Done()
	var kill, send_kill bool
EndLoop:
	for {
		select {
		case wb, wbOk := <-wbCh: //异步的写都在这里缓冲
			//channel closed
			if !wbOk {
				log4go.Debug("closed wbCh ...")
				break EndLoop
			}
			if kill && !send_kill {
				//log4go.Info("%s 🔪  %v,%v", uuid, kill, send_kill)
				wb := writeBufArgs{cmd: "KILL"}
				self.wbCh <- wb
				send_kill = true
			}
			if wb.cmd == "KILL" {
				if wbTotal > 1 {
					atomic.AddInt32(&wbTotal, -1)
					//自杀吧
					break EndLoop
				} else {
					log4go.Info("%s ⌛️  %d", uuid, wbTotal)
				}
			} else {
				log4go.Debug("handler : %s %s", wb.cmd, wb.args)
				if wb.resp == nil {
					self.executeDirect(wb.cmd, wb.args...)
				} else {
					wb.resp <- self.executeDirect(wb.cmd, wb.args...)
				}
			}
		case <-wbSyncCh:
			kill = true
		case <-time.After(time.Duration(2 * time.Second)):
			if kill {
				atomic.AddInt32(&wbTotal, -1)
				break EndLoop
			}
		}
	}
}

func (self *SaraDatabase) wbfConsumer() {
	consumerTotal := 20
	if self.PoolSize > 200 {
		consumerTotal = self.PoolSize / 10
	}
	wbSyncChList := make([]chan int, 0)
	wbSyncChList4s := make([]chan int, 0)
	self.wg.Add(consumerTotal * 2)
	log4go.Info("write buffer started ; total consume [%d]", consumerTotal*2)
	for i := 0; i < consumerTotal; i++ {
		wbSyncCh := make(chan int)
		go self.wbfConsumerWorker(wbSyncCh, self.wbCh, int32(consumerTotal))
		wbSyncChList = append(wbSyncChList, wbSyncCh)
	}
	for i := 0; i < consumerTotal; i++ {
		wbSyncCh4s := make(chan int)
		go self.wbfConsumerWorker(wbSyncCh4s, self.wbCh4s, int32(consumerTotal))
		wbSyncChList4s = append(wbSyncChList4s, wbSyncCh4s)
	}
	self.wbSyncChList = wbSyncChList
	self.wbSyncChList4s = wbSyncChList4s
}

func (self *SaraDatabase) getRedisClient(k string) (r *redis.Client) {
	switch self.model {
	case MODEL_SINGLE:
		r, _ = self.p_cli.Get()
	case MODEL_CLUSTER:
		r, _ = self.c_cli.GetForKey(k)
	default:
		r = nil
	}
	return
}

func (self *SaraDatabase) ResetExpire(key []byte, ex int) (t int, err error) {
	//if r := self.execute4s_sync("Expire", key, ex); r.Err != nil {
	if r := self.executeDirect("Expire", key, ex); r.Err != nil {
		t, err = 0, r.Err
	} else {
		if i, e := r.Int(); e != nil {
			t, err = 0, e
		} else if i == 0 {
			t, err = 0, errors.New("key_not_found")
		} else {
			t = i
		}
	}
	return
}
func (self *SaraDatabase) Put(key []byte, value []byte) error {
	if self.is4s(key) {
		if r := self.execute4s_sync("SET", key, value); r != nil && r.Err != nil {
			return r.Err
		}
	} else {
		self.execute("SET", key, value)
	}
	return nil
}
func (self *SaraDatabase) PutEx(key []byte, value []byte, ex int) error {
	if self.is4s(key) {
		if r := self.execute4s_sync("SETEX", key, ex, value); r != nil && r.Err != nil {
			return r.Err
		}
	} else {
		self.execute("SETEX", key, ex, value)
	}
	return nil
}

func (self *SaraDatabase) Get(key []byte) ([]byte, error) {
	r := self.executeDirect("GET", key)
	return r.Bytes()
}
func (self *SaraDatabase) Delete(key []byte) error {
	var r *redis.Resp
	if self.is4s(key) {
		r = self.execute4s("DEL", key)
	} else {
		r = self.execute("DEL", key)
	}
	if r != nil && r.Err != nil {
		return r.Err
	}
	return nil
}
func (self *SaraDatabase) PutExWithIdx(idx, key, value []byte, ex int) error {
	idx = append(idx, IDX_SUFFIX...)
	if self.is4s(key) {
		r := self.execute4s_sync("HSET", idx, key, value)
		if r != nil && r.Err != nil {
			return r.Err
		}
		if ex > 0 {
			r = self.execute4s_sync("SETEX", key, ex, value)
		} else {
			r = self.execute4s_sync("SET", key, value)
		}
		if r != nil && r.Err != nil {
			return r.Err
		}
	} else {
		defer func() {
			if e := recover(); e != nil {
				log4go.Error(e)
			}
		}()
		self.execute("HSET", idx, key, value)
		if ex > 0 {
			self.execute("SETEX", key, ex, value)
		} else {
			self.execute("SET", key, value)
		}
	}
	return nil
}
func (self *SaraDatabase) DeleteByIdx(idx []byte) error {
	idx = append(idx, IDX_SUFFIX...)
	if ids, err := self.executeDirect("HKEYS", idx).ListBytes(); err != nil {
		return err
	} else {
		for _, k := range ids {
			self.Delete(k)
		}
	}
	self.Delete(idx)
	return nil
}

func (self *SaraDatabase) DeleteByIdxKey(idx, key []byte) error {
	idx = append(idx, IDX_SUFFIX...)
	if self.is4s(key) {
		self.execute4s("HDEL", idx, key)
	} else {
		self.execute("HDEL", idx, key)
	}
	self.Delete(key)
	return nil
}

func (self *SaraDatabase) CountByIdx(idx []byte) (int, error) {
	idx = append(idx, IDX_SUFFIX...)
	return self.executeDirect("HLEN", idx).Int()
}

func (self *SaraDatabase) GetByIdx(idx []byte) ([][]byte, error) {
	idx = append(idx, IDX_SUFFIX...)
	log4go.Debug("idx=> %s", idx)
	if vals, err := self.executeDirect("HVALS", idx).ListBytes(); err != nil {
		return nil, err
	} else {
		return vals, nil
	}
	return nil, errors.New("empty")
}

func (self *SaraDatabase) Close() {
	self.stop <- struct{}{}
	log4go.Info("wait_close_db")
	//非 session 的缓冲区直接关闭，丢掉所有未处理消息,主要同步处理 session
	//暴力关闭
	close(self.wbCh)
	//for i, wbSyncCh := range self.wbSyncChList {
	//	wbSyncCh <- i
	//	log4go.Info("👷  consumer %d closed... (total:%d)", i, len(self.wbSyncChList))
	//}
	//安全关闭
	for i, wbSyncCh4s := range self.wbSyncChList4s {
		wbSyncCh4s <- i
		log4go.Info("👔  consumer4s %d closed... (total:%d)", i, len(self.wbSyncChList4s))
	}
	self.wg.Wait()
	switch self.model {
	case MODEL_SINGLE:
		self.p_cli.Empty()
	case MODEL_CLUSTER:
		self.c_cli.Close()
	}
	log4go.Info("success_close_db")
}

func (self *SaraDatabase) keepalive() {
	k := fmt.Sprintf("keep_%s", uuid.Rand().Hex())
over:
	for {
		select {
		case <-self.stop:
			log4go.Debug("♨️  kill keeplive thread.")
			break over
		default:
			r := self.executeDirect("SETEX", k, 4, 1)
			if self.tl {
				log4go.Debug("🐎  %s", r.String())
			}
			time.Sleep(4 * time.Second)
		}
	}
}

//是否为一个 session 相关的操作; 以 session_ 开头的 key
func (self *SaraDatabase) is4s(key []byte) bool {
	if i := bytes.Index(key, []byte(SESSION_PERFIX)); i == 0 {
		return true
	}
	return false
}

func (self *SaraDatabase) initDb() error {
	if p, err := pool.New("tcp", self.Addr, self.PoolSize); err == nil {
		self.p_cli = p
		self.model = MODEL_SINGLE
		go self.keepalive()
		go self.wbfConsumer()
		return nil
	} else {
		return err
	}
}

func (self *SaraDatabase) initClusterDb() error {
	opts := cluster.Opts{
		Addr:     self.Addr,
		PoolSize: self.PoolSize,
	}
	if c, err := cluster.NewWithOpts(opts); err == nil {
		self.c_cli = c
		self.model = MODEL_CLUSTER
		go self.keepalive()
		go self.wbfConsumer()
		return nil
	} else {
		return err
	}
}
func (self *SaraDatabase) executeDirect(cmd string, args ...interface{}) (r *redis.Resp) {
	//log4go.Debug("%s: %s", cmd, args)
	switch self.model {
	case MODEL_SINGLE:
		r = self.p_cli.Cmd(cmd, args...)
	case MODEL_CLUSTER:
		r = self.c_cli.Cmd(cmd, args...)
	}
	return
}

func (self *SaraDatabase) execute(cmd string, args ...interface{}) *redis.Resp {
	self.wbCh <- writeBufArgs{
		cmd:  cmd,
		args: args,
		resp: nil,
	}
	return nil
}

//牺牲性能，来保证 session 中关键操作的稳定性
func (self *SaraDatabase) execute4s_sync(cmd string, args ...interface{}) *redis.Resp {
	respCh := make(chan *redis.Resp)
	log4go.Debug("sync write : %s %s", cmd, args)
	self.wbCh4s <- writeBufArgs{
		cmd:  cmd,
		args: args,
		resp: respCh,
	}
	// TODO timeout process
	r := <-respCh
	return r
}

//针对session 的操作，放到一个单独的通道上执行
func (self *SaraDatabase) execute4s(cmd string, args ...interface{}) *redis.Resp {
	log4go.Debug("async write : %s %s", cmd, args)
	self.wbCh4s <- writeBufArgs{
		cmd:  cmd,
		args: args,
		resp: nil,
	}
	return nil
}

func (self *SaraDatabase) showTestLog() {
	self.tl = true
}

func (self *SaraDatabase) GenDataChannel(name string) (dc DataChannel) {
	var db *SaraDatabase
	switch self.model {
	case MODEL_SINGLE:
		db, _ = NewDatabase(self.Addr, 5)
	case MODEL_CLUSTER:
		db, _ = NewClusterDatabase(self.Addr, 5)
	}
	sub := db.getRedisClient(name)
	pub := db.getRedisClient(name)
	dc = newChannel(name, sub, pub)
	return
}

func NewDatabase(addr string, poolSize int) (*SaraDatabase, error) {
	c := &SaraDatabase{
		Addr:     addr,
		PoolSize: poolSize,
		stop:     make(chan struct{}),
		wbCh:     make(chan writeBufArgs, poolSize*2),
		wbCh4s:   make(chan writeBufArgs, poolSize*2),
		wg:       new(sync.WaitGroup),
	}
	if err := c.initDb(); err != nil {
		return nil, err
	}
	return c, nil
}

func NewClusterDatabase(addr string, poolSize int) (*SaraDatabase, error) {
	c := &SaraDatabase{
		Addr:     addr,
		PoolSize: poolSize,
		stop:     make(chan struct{}),
		wbCh:     make(chan writeBufArgs, poolSize*2),
		wbCh4s:   make(chan writeBufArgs, poolSize*2),
		wg:       new(sync.WaitGroup),
	}
	if err := c.initClusterDb(); err != nil {
		return nil, err
	}
	return c, nil
}
