package server

import (
	"testing"
	"time"
)

func TestAcceptsConnections(t *testing.T) {
	server := NewServer()
    defer server.Close()

	go server.Listen("0.0.0.0:8080")

    time.Sleep(100 * time.Millisecond)

	client := NewClient()
	err := client.Connect("0.0.0.0:8080")

	if err != nil {
		t.Errorf("Expected connection, got error: %v", err)
	}
}

func TestClosesServer(t *testing.T) {
	server := NewServer()
	go server.Listen("0.0.0.0:8080")

    time.Sleep(100 * time.Millisecond)

    server.Close()

	client := NewClient()
	err := client.Connect("0.0.0.0:8080")

	if err == nil {
		t.Errorf("Expected error, got connection")
	}
}
