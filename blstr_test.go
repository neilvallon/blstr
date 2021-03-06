package blstr

import (
	"bytes"
	"sync"
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
	if n := bb.Flood(id1, msg); n != 0 {
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
	if n := bb.Flood(id1, msg); n != 1 {
		t.Error("send should return number of subscribers that where skipped")
	}

	// ch2 full
	if n := bb.Flood(id1, msg); n != 2 {
		t.Error("send should return number of subscribers that where skipped")
	}
}

func TestUnicast(t *testing.T) {
	id1, id2, id3 := 1, 2, 3
	ch1, ch2, ch3 := make(chan []byte, 1), make(chan []byte, 1), make(chan []byte, 1)
	msg1, msg2, msg3 := []byte("msg1"), []byte("msg2"), []byte("msg3")

	bb := New()

	bb.Subscribe(id1, ch1)
	bb.Subscribe(id2, ch2)
	bb.Subscribe(id3, ch3)

	if len(bb.subscribers) != 3 {
		t.Fatal("could not initialize subscribers for test")
	}

	if err := bb.Send(id1, msg1); err != nil {
		t.Error("first send should succeed")
	} else if len(ch1) != 1 || len(ch2) != 0 || len(ch3) != 0 {
		t.Error("only one subscriber should receive message")
	}

	if err := bb.Send(id2, msg2); err != nil {
		t.Error("second send should succeed")
	} else if len(ch1) != 1 || len(ch2) != 1 || len(ch3) != 0 {
		t.Error("only one subscriber should receive message")
	}

	if err := bb.Send(id3, msg3); err != nil {
		t.Error("third send should succeed")
	} else if len(ch1) != 1 || len(ch2) != 1 || len(ch3) != 1 {
		t.Error("only one subscriber should receive message")
	}

	if err := bb.Send(id1, msg1); err == nil {
		t.Error("send should error on non-listening subscriber")
	}

	if bytes.Compare(<-ch1, msg1) != 0 ||
		bytes.Compare(<-ch2, msg2) != 0 ||
		bytes.Compare(<-ch3, msg3) != 0 {
		t.Error("incorrect messages received")
	}

	if err := bb.Send(4, msg1); err == nil {
		t.Error("send should error for non-existent subscriber")
	}
}

func TestCount(t *testing.T) {
	bb := New()

	for i := 1; i <= 10; i++ {
		bb.Subscribe(i, make(chan []byte))
		if n := bb.Count(); n != i {
			t.Log("count returns incorrect subscriber number")
			t.Errorf("got: %d - expected: %d", n, i)
		}
	}
}

func TestReset(t *testing.T) {
	bb := New()

	for i := 1; i <= 10; i++ {
		bb.Subscribe(i, make(chan []byte))
	}

	if len(bb.subscribers) != 10 {
		t.Fatal("10 subscribers should exist on hub to start test")
	}

	bb.Reset()
	if len(bb.subscribers) != 0 {
		t.Fatal("hub should have no subscribers after reset")
	}
}

func TestLockRaces(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()

	hub := New()
	defer hub.Reset()

	caller := func(id int) {
		defer wg.Done()

		ch := make(chan []byte, 5)
		hub.Subscribe(id, ch)
		hub.Count()
		hub.Send(id-1, []byte("test1"))
		hub.Flood(id, []byte("test2"))
		hub.Unsubscribe(id)
	}

	wg.Add(2)
	go caller(1)
	go caller(2)
}

func BenchmarkFloodThroughput_1(b *testing.B)      { testFloodThroughput(1, b.N) }
func BenchmarkFloodThroughput_10(b *testing.B)     { testFloodThroughput(10, b.N) }
func BenchmarkFloodThroughput_100(b *testing.B)    { testFloodThroughput(100, b.N) }
func BenchmarkFloodThroughput_1000(b *testing.B)   { testFloodThroughput(1000, b.N) }
func BenchmarkFloodThroughput_10000(b *testing.B)  { testFloodThroughput(10000, b.N) }
func BenchmarkFloodThroughput_100000(b *testing.B) { testFloodThroughput(100000, b.N) }

func testFloodThroughput(subCount, msgCount int) {
	hub := New()

	var drainWG sync.WaitGroup
	defer drainWG.Wait()

	// make subscription
	for i := 0; i < subCount; i++ {
		// channels are buffered to the number of sends
		// test requires all sends to succeed to pass and this is
		// the only way to not skip any subscribers.
		// Another test may kill users at the end and ignore
		// skipps for a more 'real world' test.
		ch := make(chan []byte, msgCount)
		if err := hub.Subscribe(i, ch); err != nil {
			panic(err)
		}

		// Start receiver
		drainWG.Add(1)
		go func(ch chan []byte) {
			defer drainWG.Done()
			for n := 0; n < msgCount; n++ {
				<-ch
			}
		}(ch)
	}

	msg := []byte("This is a string that doesn't matter since a slice is what's being sent not the array.")
	for i := 0; i < msgCount; i++ {
		go hub.Flood(-1, msg)
	}
}
