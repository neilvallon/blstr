package blstr

import (
	"bytes"
	"testing"
)

func TestSubscribe(t *testing.T) {
	id1, id2 := 1, 2
	ch1, ch2 := make(chan []byte), make(chan []byte)

	bb := New()

	if len(bb.subscribers) != 0 {
		t.Error("new ByteBroadcaster should not have any subscribers")
	}

	if err := bb.Subscribe(id1, ch1); err != nil || len(bb.subscribers) != 1 {
		t.Error("first subscription was not made")
	} else if _, ok := bb.subscribers[id1]; !ok {
		t.Error("first subscriber id incorrect")
	}

	if err := bb.Subscribe(id2, ch2); err != nil || len(bb.subscribers) != 2 {
		t.Error("second subscription was not made")
	} else if _, ok := bb.subscribers[id2]; !ok {
		t.Error("second subscriber id incorrect")
	}

	if err := bb.Subscribe(id1, ch1); err == nil || len(bb.subscribers) != 2 {
		t.Error("should not allow duplicate subscriptions")
	}
}

func TestUnsubscribe(t *testing.T) {
	id1, id2 := 1, 2
	ch1, ch2 := make(chan []byte), make(chan []byte)

	bb := New()

	bb.Subscribe(id1, ch1)
	bb.Subscribe(id2, ch2)

	if len(bb.subscribers) != 2 {
		t.Fatal("could not initialize subscribers for test")
	}

	bb.Unsubscribe(id1)
	if _, ok := bb.subscribers[id1]; ok || len(bb.subscribers) != 1 {
		t.Error("first subscription was not removed")
	}

	bb.Unsubscribe(id2)
	if _, ok := bb.subscribers[id2]; ok || len(bb.subscribers) != 0 {
		t.Error("second subscription was not removed")
	}
}

func TestBroadcast(t *testing.T) {
	id1, id2, id3 := 1, 2, 3
	ch1, ch2, ch3 := make(chan []byte, 1), make(chan []byte, 1), make(chan []byte, 1)

	bb := New()

	bb.Subscribe(id1, ch1)
	bb.Subscribe(id2, ch2)
	bb.Subscribe(id3, ch3)

	if len(bb.subscribers) != 3 {
		t.Fatal("could not initialize subscribers for test")
	}

	msg := []byte("Hello, World!")
	if n := bb.Send(id1, msg); n != 0 {
		t.Error("zero subscribers skipped if all are listening")
	}

	if len(ch1) != 0 {
		t.Error("broadcast should not return to sender")
	}

	if len(ch2) != 1 {
		t.Error("second subscriber should receive the message")
	} else if bytes.Compare(<-ch2, msg) != 0 {
		t.Error("did not receive same message that was sent")
	}

	// ch3 full
	if n := bb.Send(id1, msg); n != 1 {
		t.Error("send should return number of subscribers that where skipped")
	}

	// ch2 full
	if n := bb.Send(id1, msg); n != 2 {
		t.Error("send should return number of subscribers that where skipped")
	}
}
