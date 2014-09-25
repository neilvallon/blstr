package blstr

import (
	"errors"
	"sync"
)

type Subscribable interface {
	Subscribe(int, chan<- []byte) error
	Unsubscribe(int)
}

type Broadcaster interface {
	Send(int, []byte) int
}

type Hub interface {
	Subscribable
	Broadcaster
}

type ByteHub struct {
	subscribers map[int]chan<- []byte
	sync.RWMutex
}

func New() *ByteHub {
	return &ByteHub{subscribers: make(map[int]chan<- []byte)}
}

func (bh *ByteHub) Subscribe(id int, ch chan<- []byte) error {
	bh.Lock()
	defer bh.Unlock()

	if _, ok := bh.subscribers[id]; ok {
		return errors.New("subscriber already exists")
	}

	bh.subscribers[id] = ch

	return nil
}

func (bh *ByteHub) Unsubscribe(id int) {
	bh.Lock()
	defer bh.Unlock()

	delete(bh.subscribers, id)
}

func (bh *ByteHub) Send(sender int, msg []byte) (skipped int) {
	bh.RLock()
	defer bh.RUnlock()

	for id, ch := range bh.subscribers {
		if id != sender {
			select {
			case ch <- msg:
			default:
				skipped++
			}
		}
	}

	return
}
