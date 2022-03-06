package server

import (
	"testing"
	"time"

	"example.com/game/client/client"
)

func TestAcceptsConnections(t *testing.T) {
	server := NewServer([]EventHandler{})
	defer server.Close()

	go server.Listen("0.0.0.0:8080")

	time.Sleep(100 * time.Millisecond)

	client := client.NewClient()
	err := client.Connect("0.0.0.0:8080")

	if err != nil {
		t.Errorf("Expected connection, got error: %v", err)
	}
}

func TestCloseServer(t *testing.T) {
	client := client.NewClient()
	err := client.Connect("0.0.0.0:8080")

	if err == nil {
		t.Errorf("Expected error, got connection")
	}
}
