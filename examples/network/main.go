package main

import (
	"math"
	"time"

	"log"

	"github.com/NeilVallon/blstr"
)

type Node struct {
	id  int
	hub *blstr.ByteHub
}

func BuildTree(p *Node, layer, fanout int, accum []*Node) {
	if layer == 1 {
		accum[p.id] = p
		return
	}

	for i := 0; i < fanout; i++ {
		n := &Node{
			id:  p.id*fanout + i + 1,
			hub: blstr.New(),
		}

		go connect(p, n)
		go connect(n, p)

		accum[n.id] = n

		BuildTree(n, layer-1, fanout, accum)
	}
}

func connect(n1, n2 *Node) {
	ch := make(chan []byte, 1)
	n2.hub.Subscribe(n1.id, ch)

	for msg := range ch {
		log.Printf("node %d forwarding to node %d", n2.id, n1.id)
		n1.hub.Flood(n2.id, msg)
	}
}

//      0
//    /   \
//   1     2
//  / \   / \
// *   4 5   *
//
// main builds a tree of hubs and atempts to send a message between
// the two furthest leaf nodes.
func main() {
	layers, fanout := 4, 4

	totalNodes := (int(math.Pow(float64(fanout), float64(layers))) - 1) / (fanout - 1)

	accum := make([]*Node, totalNodes)
	accum[0] = &Node{hub: blstr.New()}

	BuildTree(accum[0], layers, fanout, accum)

	// Give time for subscriptions to start up
	time.Sleep(50 * time.Millisecond)

	// Get bottom layer of tree
	start := (int(math.Pow(float64(fanout), float64(layers-1))) - 1) / (fanout - 1)
	end := totalNodes
	edge := accum[start:end]

	// Find far left and far right nodes
	left, right := edge[0], edge[len(edge)-1]
	log.Printf("sending message from node %d to %d", right.id, left.id)

	// Subscribe to left
	ch := make(chan []byte, 1)
	left.hub.Subscribe(-1, ch)

	// Send on right
	right.hub.Flood(-1, []byte("Hello, World!"))

	// Wait for propogation and print
	msg := <-ch
	log.Printf("%s\n", msg)
}
