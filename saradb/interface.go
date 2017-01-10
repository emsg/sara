// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package saradb

type Database interface {
	Put(key []byte, value []byte) error
	PutEx(key []byte, value []byte, ex int) error
	Get(key []byte) ([]byte, error)
	PutExWithIdx(idx, key, value []byte, ex int) error
	GetByIdx(idx []byte) ([][]byte, error)

	DeleteByIdx(idx []byte) error
	DeleteByIdxKey(idx, key []byte) error

	Delete(key []byte) error
	GenDataChannel(name string) DataChannel
	Close()
}

type SubHandler func(message string)
type DataChannel interface {
	GetChannel() string
	Publish(channel, message string) error
	Subscribe(handler SubHandler)
}
