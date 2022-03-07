package server

import (
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestPushEmpty(t *testing.T) {
	queue := &Queue{
		mut:     &sync.Mutex{},
		sockets: make(map[*websocket.Conn]*Node),
	}

	socket := &websocket.Conn{}
	queue.Push(socket)

	if queue.Head == nil {
		t.Error("Expected head to have node")
	}
	if queue.Head.Next != nil {
		t.Error("Expected head to not have next")
	}
	if queue.Head.Prev != nil {
		t.Error("Expected head to not have previous")
	}
	if queue.Tail != queue.Head {
		t.Error("Expected head and tail to point to the same node")
	}
	if queue.Head.Socket != socket {
		t.Error("Expected head to have socket")
	}
	if len(queue.sockets) != 1 {
		t.Errorf("Expected sockets to have count of 1, got %d", len(queue.sockets))
	}
}

func TestPushNotEmpty(t *testing.T) {
	queue := &Queue{
		mut:     &sync.Mutex{},
		sockets: make(map[*websocket.Conn]*Node),
	}

	first := &websocket.Conn{}
	last := &websocket.Conn{}

	queue.Push(first)
	queue.Push(last)

	if queue.Head.Socket != first {
		t.Error("Expected head to point to first")
	}
	if queue.Tail.Socket != last {
		t.Error("Expected tail to point to last")
	}
	if queue.Head.Next != queue.Tail {
		t.Error("Expected head to point to tail")
	}
	if queue.Tail.Prev != queue.Head {
		t.Error("Expected tail to point to head")
	}
	if len(queue.sockets) != 2 {
		t.Errorf("Expected sockets to have count of 2, got %d", len(queue.sockets))
	}
}

func TestPushMultiple(t *testing.T) {
	queue := &Queue{
		mut:     &sync.Mutex{},
		sockets: make(map[*websocket.Conn]*Node),
	}

	first := &websocket.Conn{}
	middle := &websocket.Conn{}
	last := &websocket.Conn{}

	queue.Push(first)
	queue.Push(middle)
	queue.Push(last)

	if queue.Head.Socket != first {
		t.Error("Expected head to point to first")
	}
	if queue.Head.Next.Socket != middle {
		t.Error("Expected middle to point to middle")
	}
	if queue.Tail.Socket != last {
		t.Error("Expected tail to point to last")
	}
	if queue.Head.Next.Next != queue.Tail {
		t.Error("Expected middle next to point to tail")
	}
	if queue.Tail.Prev.Prev != queue.Head {
		t.Error("Expected middle prev to point to head")
	}
	if len(queue.sockets) != 3 {
		t.Errorf("Expected sockets to have count of 3, got %d", len(queue.sockets))
	}
	if queue.Tail.Next != nil {
		t.Error("Expected tail next to point to nil")
	}
}

func TestPopEmpty(t *testing.T) {
	queue := &Queue{
		mut:     &sync.Mutex{},
		sockets: make(map[*websocket.Conn]*Node),
	}

	node := queue.Pop()

	if node != nil {
		t.Errorf("Expected pop on empty to return nil, got %v", node)
	}
	if len(queue.sockets) != 0 {
		t.Errorf("Expected sockets to have count of 0, got %d", len(queue.sockets))
	}
}

func TestPopSingle(t *testing.T) {
	queue := &Queue{
		mut:     &sync.Mutex{},
		sockets: make(map[*websocket.Conn]*Node),
	}

	socket := &websocket.Conn{}
	queue.Push(socket)

	node := queue.Pop()

	if node == nil {
		t.Error("Expected pop to return socket, got nil")
	}
	if node != socket {
		t.Error("Expected pop to return added socket")
	}
	if queue.Head != nil {
		t.Error("Expected empty head")
	}
	if queue.Tail != nil {
		t.Error("Expected empty tail")
	}
	if len(queue.sockets) != 0 {
		t.Errorf("Expected sockets to have count of 0, got %d", len(queue.sockets))
	}
}

func TestPopPair(t *testing.T) {
	queue := &Queue{
		mut:     &sync.Mutex{},
		sockets: make(map[*websocket.Conn]*Node),
	}

	first := &websocket.Conn{}
	last := &websocket.Conn{}

	queue.Push(first)
	queue.Push(last)

	node := queue.Pop()

	if node != first {
		t.Error("Expected pop to return first")
	}
	if queue.Head.Socket != last {
		t.Error("Expected head to point to last")
	}
	if queue.Head != queue.Tail {
		t.Error("Expected head and tail to point to the same node")
	}
	if queue.Head.Prev != nil {
		t.Error("Expected head prev to point to nil")
	}
	if queue.Head.Next != nil {
		t.Error("Expected head next to point to nil")
	}
	if queue.Tail.Next != nil {
		t.Error("Expected tail next to point to nil")
	}
	if queue.Tail.Prev != nil {
		t.Error("Expected tail prev to point to nil")
	}
}

func TestRemoveMiddle(t *testing.T) {
	queue := &Queue{
		mut:     &sync.Mutex{},
		sockets: make(map[*websocket.Conn]*Node),
	}

	first := &websocket.Conn{}
	middle := &websocket.Conn{}
	last := &websocket.Conn{}

	queue.Push(first)
	queue.Push(middle)
	queue.Push(last)

	queue.Remove(middle)

	if queue.Head.Next != queue.Tail {
		t.Error("Expected head next to point to tail")
	}
	if queue.Tail.Prev != queue.Head {
		t.Error("Expected tail prev to point to head")
	}
}

func TestRemoveLast(t *testing.T) {
	queue := &Queue{
		mut:     &sync.Mutex{},
		sockets: make(map[*websocket.Conn]*Node),
	}

	first := &websocket.Conn{}
	middle := &websocket.Conn{}
	last := &websocket.Conn{}

	queue.Push(first)
	queue.Push(middle)
	queue.Push(last)

	queue.Remove(last)

	if queue.Tail.Next != nil {
		t.Error("Expected tail next to point to nil")
	}
	if queue.Tail.Socket != middle {
		t.Error("Expected tail to be middle")
	}
	if queue.Head.Next != queue.Tail {
		t.Error("Expected head next to point to tail")
	}
	if queue.Head.Socket != first {
		t.Error("Expected head to point to first")
	}
}

func TestQueueCommand(t *testing.T) {
	queueManager := NewQueueManager()
	server := NewServer([]EventHandler{
		queueManager,
	})
	defer server.Close()

	go server.Listen("0.0.0.0:8080")

	time.Sleep(time.Millisecond)

	c := NewTestClient()
	response := c.QueueUp()

	if response.Type != "wait_for_match" {
		t.Errorf("Expected \"wait_for_match\", got \"%s\"", response.Type)
	}
}

func TestQueuesUser(t *testing.T) {
	queueManager := NewQueueManager()
	server := NewServer([]EventHandler{
		queueManager,
	})
	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(time.Millisecond)

	c := NewTestClient()
	c.QueueUp()

	if queueManager.queue.Count() != 1 {
		t.Errorf("Expected queue to have one, got %d", queueManager.queue.Count())
	}
}

type FakeMatchMaker struct {
	Invoked int
}

func (m *FakeMatchMaker) Process(event Event, server *Server) {
	if event.Type == "match_found" {
		m.Invoked += 1
	}
}

func TestDispatchesMatchFound(t *testing.T) {
	queueManager := NewQueueManager()
	fakeMaker := &FakeMatchMaker{0}

	server := NewServer([]EventHandler{
		queueManager,
		fakeMaker,
	})
	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(time.Millisecond)

	c1 := NewTestClient()
	c2 := NewTestClient()

	c1.QueueUp()
	c2.QueueUp()

	if fakeMaker.Invoked != 1 {
		t.Errorf("Expected match found to be dispatched, got %d", fakeMaker.Invoked)
	}
}

func TestDisconnectRemovesFromQueue(t *testing.T) {
	queueManager := NewQueueManager()

	server := NewServer([]EventHandler{
		queueManager,
	})
	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(time.Millisecond)

	c1 := NewTestClient()
	c1.QueueUp()

	c1.Client.Close()

	// TODO: How to not do this?
	time.Sleep(time.Millisecond)

	if len(queueManager.queue.sockets) != 0 {
		t.Errorf("Expected queue to have count 0, got %d", len(queueManager.queue.sockets))
	}
}
