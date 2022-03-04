package server

import (
	"testing"
	"time"
)

func TestQueueCommand(t *testing.T) {
	queueManager := NewQueueManager()
	server := NewServer([]EventHandler{
		queueManager,
	})
	defer server.Close()

	go server.Listen("0.0.0.0:8080")

	time.Sleep(100 * time.Millisecond)

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

	time.Sleep(100 * time.Millisecond)

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

	time.Sleep(100 * time.Millisecond)

	c1 := NewTestClient()
	c2 := NewTestClient()

	c1.QueueUp()
	c2.QueueUp()

	if fakeMaker.Invoked != 1 {
		t.Errorf("Expected match found to be dispatched, got %d", fakeMaker.Invoked)
	}
}
