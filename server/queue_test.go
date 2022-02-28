package server

import (
	"testing"
	"time"

	"example.com/game/client"
	"example.com/game/common"
)

func TestQueueCommand(t *testing.T) {
	queueManager := NewQueueManager()
	server := NewServer([]EventHandler{
		queueManager,
	})
	defer server.Close()

	go server.Listen("0.0.0.0:8080")

	time.Sleep(100 * time.Millisecond)

	c := client.NewClient()
	c.Connect("0.0.0.0:8080")

	c.Send(common.QueueUp())

	response := <-c.Incoming

	if response.Type != "wait" {
		t.Errorf("Expected \"wait\", got \"%s\"", response.Type)
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

	c := client.NewClient()
	c.Connect("0.0.0.0:8080")

	c.Send(common.QueueUp())

	time.Sleep(100 * time.Millisecond)

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

	c1 := client.NewClient()
	c2 := client.NewClient()

	c1.Connect("0.0.0.0:8080")
	c2.Connect("0.0.0.0:8080")

	c1.Send(common.QueueUp())
	c2.Send(common.QueueUp())

	time.Sleep(100 * time.Millisecond)

	if fakeMaker.Invoked != 1 {
		t.Errorf("Expected match found to be dispatched, got %d", fakeMaker.Invoked)
	}
}
