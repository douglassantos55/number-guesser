package server

import (
	"testing"
	"time"

	"example.com/game/client"
	"example.com/game/common"
)

func TestQueueCommand(t *testing.T) {
	server := NewServer()
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
