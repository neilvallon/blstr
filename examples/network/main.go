package main

import (
	"math"
	"sync"

	"log"

	"github.com/NeilVallon/blstr"
)

type Node struct {
	id  int
	hub *blstr.ByteHub
}

type Layer []Node
type Tree []Layer

//      0
//    /   \
//   1     2
//  / \   / \
// *   4 5   *
//
// main builds a tree of hubs and atempts to send a message between
// the two furthest leaf nodes.
func main() {
	tree := BuildTree(4, 4)

	// Get bottom layer of tree
	edge := tree[len(tree)-1]

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

func BuildTree(layers, fanout int) Tree {
	// Make a single layer to hold all nodes of the final tree
	accum := make(Layer, totalNodes(layers, fanout))

	// Set root 0 Node
	accum[0] = Node{hub: blstr.New()}

	var wg sync.WaitGroup
	defer wg.Wait()

	buildTree(accum[0], layers, fanout, accum, &wg)

	// Form final tree by slicing out layers from the accumulator
	tree := make(Tree, layers)
	for i := 1; i <= layers; i++ {
		s, e := totalNodes(i-1, fanout), totalNodes(i, fanout)
		tree[i-1] = accum[s:e]
	}

	return tree
}

func totalNodes(layers, fanout int) int {
	return int(math.Pow(float64(fanout), float64(layers))-1) / (fanout - 1)
}

func buildTree(p Node, layers, fanout int, accum Layer, wg *sync.WaitGroup) {
	if layers == 1 {
		accum[p.id] = p
		return
	}

	for i := 0; i < fanout; i++ {
		n := Node{
			id:  p.id*fanout + i + 1,
			hub: blstr.New(),
		}

		// Forward messages between parent and new node
		wg.Add(2)
		go connect(&p, &n, wg)
		go connect(&n, &p, wg)

		accum[n.id] = n

		// make subtree bellow new node
		buildTree(n, layers-1, fanout, accum, wg)
	}
}

func connect(n1, n2 *Node, wg *sync.WaitGroup) {
	ch := make(chan []byte, 1)
	n2.hub.Subscribe(n1.id, ch)

	wg.Done()
	for msg := range ch {
		log.Printf("node %d forwarding to node %d", n2.id, n1.id)
		n1.hub.Flood(n2.id, msg)
	}
}
