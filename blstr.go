// Package blstr (blaster? blast string?) is an ill-named library
// for easily setting up best-effort broadcast communication between
// goroutines.
//
// Synchronization and consistency is traded for speed and
// predictability of send operations.
//
// The best use case is where perfect consistency may not be
// necessary and some message loss on slow receivers is more tolerable
// than choking upstream senders.
package blstr // import "vallon.me/blstr"

import (
	"errors"
	"sync"
)

// ByteHub implements a multiuser communication channel.
type ByteHub struct {
	subscribers map[int]chan<- []byte
	sync.RWMutex
}

// New returns a new ByteHub
func New() *ByteHub {
	return &ByteHub{subscribers: make(map[int]chan<- []byte)}
}

// Subscribe to be notified when broadcast events occur on the hub.
//
// Messages can be dropped if receiver stops listening, or if messages
// arrive faster than they can be consumed.
// Channels can be buffered to mitigate this to some extent.
func (bh *ByteHub) Subscribe(id int, ch chan<- []byte) error {
	bh.Lock()

	if _, ok := bh.subscribers[id]; ok {
		bh.Unlock()
		return errors.New("subscriber already exists")
	}

	bh.subscribers[id] = ch

	bh.Unlock()
	return nil
}

// Unsubscribe tries to remove the subscription by ID if it exists.
func (bh *ByteHub) Unsubscribe(id int) {
	bh.Lock()
	delete(bh.subscribers, id)
	bh.Unlock()
}

// Send provides unicast communication over the hub.
//
// An attempt is made to send a message to the specified subscriber.
// Errors are returned if the subscriber could not be found or
// is unable to receive the message.
func (bh *ByteHub) Send(to int, msg []byte) (err error) {
	bh.RLock()

	ch, ok := bh.subscribers[to]
	if !ok {
		bh.RUnlock()
		return errors.New("subscriber does not exist")
	}

	select {
	case ch <- msg:
	default:
		err = errors.New("subscriber not listening")
	}

	bh.RUnlock()
	return
}

// Flood provides broadcast communication over the hub.
//
// All listening subscribers will receive the message provided.
// The from field is usually set with the sender ID to be excluded
// from the broadcast. This allows subscribers to send without
// having to filter out their own messages.
//
// A count of skipped subscribers (excluding the sender) that are
// unable to receive the message is returned.
func (bh *ByteHub) Flood(from int, msg []byte) (skipped int) {
	bh.RLock()

	for id, ch := range bh.subscribers {
		if id != from {
			select {
			case ch <- msg:
			default:
				skipped++
			}
		}
	}

	bh.RUnlock()
	return
}

// Count returns the current number of subscriptions to the hub.
func (bh *ByteHub) Count() int {
	bh.RLock()
	l := len(bh.subscribers)
	bh.RUnlock()
	return l
}

// Reset removes all subscriptions from the hub and allows it to be reused.
func (bh *ByteHub) Reset() {
	bh.Lock()
	bh.subscribers = make(map[int]chan<- []byte)
	bh.Unlock()
}
