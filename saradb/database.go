package saradb

import (
	"errors"
	"fmt"
	"sara/utils"
	"time"

	"github.com/alecthomas/log4go"
	"github.com/golibs/uuid"
	"github.com/mediocregopher/radix.v2/cluster"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

type DBMODEL int

const (
	MODEL_CLUSTER DBMODEL = iota
	MODEL_SINGLE
)

type SaraDatabase struct {
	Addr     string
	PoolSize int
	c_cli    *cluster.Cluster
	p_cli    *pool.Pool
	model    DBMODEL
	stop     chan struct{}
	tl       bool
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
	r := self.execute("GET", key)
	return r.Bytes()
}
func (self *SaraDatabase) Delete(key []byte) error {
	self.execute("DEL", key)
	return nil
}
func (self *SaraDatabase) PutExWithIdx(idx, key, value []byte, ex int) error {
	self.execute("ZADD", idx, utils.Timestamp13(), key)
	self.execute("SETEX", key, ex, value)
	return nil
}
func (self *SaraDatabase) DeleteByIdx(idx []byte) error {
	if ids, err := self.execute("ZRANGE", idx, 0, -1).ListBytes(); err != nil {
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
	self.execute("ZREM", idx, key)
	self.Delete(key)
	return nil
}

func (self *SaraDatabase) GetByIdx(idx []byte) ([][]byte, error) {
	if ids, err := self.execute("ZRANGE", idx, 0, -1).ListBytes(); err != nil {
		return nil, err
	} else {
		var r [][]byte
		for _, k := range ids {
			if v, e := self.Get(k); e == nil && v != nil {
				r = append(r, v)
			}
		}
		return r, nil
	}
	return nil, errors.New("empty")
}

func (self *SaraDatabase) Close() {
	self.stop <- struct{}{}
	switch self.model {
	case MODEL_SINGLE:
		self.p_cli.Empty()
	case MODEL_CLUSTER:
		self.c_cli.Close()
	}
}

func (self *SaraDatabase) keepalive() {
	k := fmt.Sprintf("keep_%s", uuid.Rand().Hex())
over:
	for {
		select {
		case <-self.stop:
			log4go.Debug("♨️  shutdown cluster_database")
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
		return nil
	} else {
		return err
	}
}
func (self *SaraDatabase) execute(cmd string, args ...interface{}) (r *redis.Resp) {
	switch self.model {
	case MODEL_SINGLE:
		r = self.p_cli.Cmd(cmd, args...)
	case MODEL_CLUSTER:
		r = self.c_cli.Cmd(cmd, args...)
	}
	return
}
func (self *SaraDatabase) showTestLog() {
	self.tl = true
}

func (self *SaraDatabase) GenDataChannel(name string) (dc DataChannel) {
	var db *SaraDatabase
	switch self.model {
	case MODEL_SINGLE:
		db, _ = NewDatabase(self.Addr, 500)
	case MODEL_CLUSTER:
		db, _ = NewClusterDatabase(self.Addr, 500)
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
	}
	if err := c.initClusterDb(); err != nil {
		return nil, err
	}
	return c, nil
}
