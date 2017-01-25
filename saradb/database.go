package saradb

import (
	"errors"
	"fmt"
	//"sara/utils"
	"sync"
	"time"

	"github.com/alecthomas/log4go"
	"github.com/golibs/uuid"
	"github.com/mediocregopher/radix.v2/cluster"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

type DBMODEL int

type writeBufArgs struct {
	cmd  string
	args []interface{}
}

const (
	MODEL_CLUSTER DBMODEL = iota
	MODEL_SINGLE
)

type SaraDatabase struct {
	Addr         string
	PoolSize     int
	c_cli        *cluster.Cluster
	p_cli        *pool.Pool
	model        DBMODEL
	stop         chan struct{}
	tl           bool
	wbCh         chan writeBufArgs //异步的写操作缓冲区
	wbTotal      int               //写操作工作线程数
	wbSyncChList []chan int        //同步写操作工作线程
	wg           *sync.WaitGroup   //同步写操作工作线程
	lock         *sync.RWMutex
}

func (self *SaraDatabase) wbfConsumerWorker(wbSyncCh chan int) {
	uuid := uuid.Rand().Hex()
	defer self.wg.Done()
	var kill, send_kill bool
EndLoop:
	for {
		select {
		case wb := <-self.wbCh:
			if kill && !send_kill {
				//log4go.Info("%s 🔪  %v,%v", uuid, kill, send_kill)
				wb := writeBufArgs{cmd: "KILL"}
				self.wbCh <- wb
				send_kill = true
			}
			if wb.cmd == "KILL" {
				if self.wbTotal > 1 {
					self.lock.Lock()
					self.wbTotal -= 1
					//自杀吧
					self.lock.Unlock()
					break EndLoop
				} else {
					log4go.Info("%s ⌛️  %d", uuid, self.wbTotal)
				}
			} else {
				self.executeDirect(wb.cmd, wb.args...)
			}
		case <-wbSyncCh:
			kill = true
		case <-time.After(time.Duration(2 * time.Second)):
			if kill {
				self.lock.Lock()
				self.wbTotal -= 1
				self.lock.Unlock()
				break EndLoop
			}
		}
	}
}

func (self *SaraDatabase) wbfConsumer() {
	consumerTotal := 10
	if self.PoolSize > 200 {
		consumerTotal = self.PoolSize / 20
	}
	log4go.Info("write buffer started ; total consume [%d]", consumerTotal)
	wbSyncChList := make([]chan int, 0)
	self.wbTotal = consumerTotal
	self.wg.Add(consumerTotal)
	for i := 0; i < consumerTotal; i++ {
		wbSyncCh := make(chan int)
		go self.wbfConsumerWorker(wbSyncCh)
		wbSyncChList = append(wbSyncChList, wbSyncCh)
	}
	self.wbSyncChList = wbSyncChList
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

func (self *SaraDatabase) Put(key []byte, value []byte) error {
	self.execute("SET", key, value)
	return nil
}
func (self *SaraDatabase) PutEx(key []byte, value []byte, ex int) error {
	self.execute("SETEX", key, ex, value)
	return nil
}

func (self *SaraDatabase) Get(key []byte) ([]byte, error) {
	r := self.executeDirect("GET", key)
	return r.Bytes()
}
func (self *SaraDatabase) Delete(key []byte) error {
	r := self.execute("DEL", key)
	if r != nil && r.Err != nil {
		return r.Err
	}
	return nil
}
func (self *SaraDatabase) PutExWithIdx(idx, key, value []byte, ex int) error {
	//r := self.execute("ZADD", idx, utils.Timestamp13(), key)
	//val := fmt.Sprintf("%d%s", utils.Timestamp13(), key)
	r := self.execute("HSET", idx, key, value)
	if r != nil && r.Err != nil {
		return r.Err
	}
	if ex > 0 {
		r = self.execute("SETEX", key, ex, value)
	} else {
		r = self.execute("SET", key, value)
	}
	if r != nil && r.Err != nil {
		return r.Err
	}
	return nil
}
func (self *SaraDatabase) DeleteByIdx(idx []byte) error {
	if ids, err := self.executeDirect("HKEYS", idx).ListBytes(); err != nil {
		return err
	} else {
		for _, k := range ids {
			self.Delete(k)
		}
	}
	self.Delete(idx)
	/*
		if ids, err := self.executeDirect("ZRANGE", idx, 0, -1).ListBytes(); err != nil {
			return err
		} else {
			for _, k := range ids {
				self.Delete(k)
			}
		}
	*/
	return nil
}

func (self *SaraDatabase) DeleteByIdxKey(idx, key []byte) error {
	self.execute("HDEL", idx, key)
	//self.execute("ZREM", idx, key)
	self.Delete(key)
	return nil
}

func (self *SaraDatabase) CountByIdx(idx []byte) (int, error) {
	return self.executeDirect("HLEN", idx).Int()
}

func (self *SaraDatabase) GetByIdx(idx []byte) ([][]byte, error) {
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
	for i, wbSyncCh := range self.wbSyncChList {
		wbSyncCh <- i
		log4go.Info("👷  consumer %d closed... (total:%d)", i, len(self.wbSyncChList))
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
			r := self.execute("SETEX", k, 4, 1)
			if self.tl {
				log4go.Debug("🐎  %s", r.String())
			}
			time.Sleep(4 * time.Second)
		}
	}
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
	switch self.model {
	case MODEL_SINGLE:
		r = self.p_cli.Cmd(cmd, args...)
	case MODEL_CLUSTER:
		r = self.c_cli.Cmd(cmd, args...)
	}
	return
}
func (self *SaraDatabase) execute(cmd string, args ...interface{}) *redis.Resp {
	//TODO 通过队列进行缓冲
	wb := writeBufArgs{
		cmd:  cmd,
		args: args,
	}
	//self.wg.Add(1)
	self.wbCh <- wb
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
		wbCh:     make(chan writeBufArgs, 20000),
		wg:       new(sync.WaitGroup),
		lock:     new(sync.RWMutex),
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
		wbCh:     make(chan writeBufArgs, 20000),
		wg:       new(sync.WaitGroup),
		lock:     new(sync.RWMutex),
	}
	if err := c.initClusterDb(); err != nil {
		return nil, err
	}
	return c, nil
}
