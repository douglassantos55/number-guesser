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
