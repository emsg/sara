package utils

import "time"

func Timestamp10() int64 {
	return time.Now().Unix()
}

func Timestamp13() int64 {
	return time.Now().UnixNano() / 1000000
}

func TimeoutTime(second int) time.Time {
	return time.Now().Add(time.Second * time.Duration(second))
}
