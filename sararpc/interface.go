package sararpc

type SubHandler func(message string)
type DataChannel interface {
	GetChannel() string
	Publish(channel, message string) error
	Subscribe(handler SubHandler)
}
