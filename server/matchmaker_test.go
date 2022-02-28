package server

import (
	"testing"
	"time"

	"example.com/game/client"
	"example.com/game/common"
)

func TestMatchFound(t *testing.T) {
	matchMaker := NewMatchMaker()
	server := NewServer([]EventHandler{
		matchMaker,
	})
	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(100 * time.Millisecond)

	c1 := client.NewClient()

	c1.Connect("0.0.0.0:8080")

	c1.Send(common.Message{
		Type: "match_found",
	})

	select {
	case response := <-c1.Incoming:
		if response.Type != "confirm_match" {
			t.Errorf("Expected confirm match response, got %s", response.Type)
		}

        if int(response.Payload["matchId"].(float64)) != 1 {
			t.Errorf("Expected 1, got %d", int(response.Payload["matchId"].(float64)))
        }
	case <-time.After(time.Second):
		t.Errorf("Expected confirm match response, got timeout")
	}
}
