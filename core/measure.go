package core

import "sync/atomic"

var (
	read  int64
	write int64
)

func MeasureReadAdd(i int) {
	atomic.AddInt64(&read, int64(i))
}

func MeasureWriteAdd(i int) {
	atomic.AddInt64(&write, int64(i))
}

func MeasureReadGet() int64 {
	return atomic.LoadInt64(&read)
}
func MeasureWriteGet() int64 {
	return atomic.LoadInt64(&write)
}
