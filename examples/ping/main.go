package main

import (
	"fmt"
	"sync"

	"log"

	"github.com/NeilVallon/blstr"
)

const pings = 10

var wg, fin sync.WaitGroup

func main() {
	h := blstr.New()

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		fin.Add(1)
		go pinger(i, h)
	}

	// Wait for subscriptions to be made
	wg.Wait()

	// Seed first pinger
	h.Send(1, []byte("initial ping"))

	// Wait for pingers to end
	fin.Wait()
}

func pinger(id int, hub *blstr.ByteHub) {
	ch := make(chan []byte, pings)
	hub.Subscribe(id, ch)
	wg.Done()

	for i := 0; i < pings; i++ {
		ping := <-ch
		log.Printf("%d received %s - sending pong\n", id, ping)

		pong := []byte(fmt.Sprintf("ping #%d from %d", i, id))
		if n := hub.Flood(id, pong); n != 0 {
			log.Printf("skiped %d subscribers when sending ping #%d from %d\n", n, i, id)
		}
	}

	hub.Unsubscribe(id)
	log.Printf("pinger %d responded to %d pings and unsubscribed\n", id, pings)

	fin.Done()
}
