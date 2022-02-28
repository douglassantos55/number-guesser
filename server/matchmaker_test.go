package server

import (
	"log"
	"testing"
	"time"

	"example.com/game/client"
	"example.com/game/common"
)

type TestClient struct {
	Client *client.Client
}

func NewTestClient() *TestClient {
	testClient := &TestClient{
		Client: client.NewClient(),
	}
	testClient.Client.Connect("0.0.0.0:8080")
	return testClient
}

func (c *TestClient) GetIncoming() common.Message {
	select {
	case msg := <-c.Client.Incoming:
		return msg
	case <-time.After(time.Second):
		log.Fatal("Timeout")
	}
	// should never reach this
	return common.Message{}
}

func (c *TestClient) QueueUp() common.Message {
	c.Client.Send(common.Message{
		Type: "queue_up",
	})
	return c.GetIncoming()
}

func (c *TestClient) AcceptMatch(match int) common.Message {
	c.Client.Send(common.Message{
		Type: "match_confirmed",
		Payload: map[string]interface{}{
			"matchId": match,
		},
	})
	time.Sleep(100 * time.Millisecond)
	return c.GetIncoming()
}

func TestMatchFound(t *testing.T) {
	server := NewServer([]EventHandler{
		NewQueueManager(),
		NewMatchMaker(),
	})

	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(100 * time.Millisecond)

	c1 := NewTestClient()
	c2 := NewTestClient()

	c1.QueueUp()
	c2.QueueUp()

	res1 := c1.GetIncoming()
	if res1.Type != "confirm_match" {
		t.Errorf("Expected confirm match response, got %s", res1.Type)
	}

	if int(res1.Payload["matchId"].(float64)) != 1 {
		t.Errorf("Expected 1, got %d", int(res1.Payload["matchId"].(float64)))
	}

	res2 := c2.GetIncoming()
	if res2.Type != "confirm_match" {
		t.Errorf("Expected confirm match response, got %s", res2.Type)
	}

	if int(res2.Payload["matchId"].(float64)) != 1 {
		t.Errorf("Expected 1, got %d", int(res2.Payload["matchId"].(float64)))
	}
}

func TestConfirmsMatch(t *testing.T) {
	matchMaker := NewMatchMaker()
	server := NewServer([]EventHandler{
		NewQueueManager(),
		matchMaker,
	})

	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(100 * time.Millisecond)

	c1 := NewTestClient()
	c2 := NewTestClient()

	c1.QueueUp()
	c2.QueueUp()

	c1.AcceptMatch(1)
	if matchMaker.Confirmed(1) != 1 {
		t.Errorf("Expected 1 confirmed, got %d", matchMaker.Confirmed(1))
	}

	c2.AcceptMatch(1)
	if matchMaker.HasMatch(1) {
		t.Errorf("Expected match to be resolved %v", matchMaker.matches)
	}

	res1 := c1.GetIncoming()
	if res1.Type != "game_start" {
		t.Errorf("Expected \"game_start\", got %s", res1.Type)
	}
	res2 := c2.GetIncoming()
	if res2.Type != "game_start" {
		t.Errorf("Expected \"game_start\", got %s", res2.Type)
	}
}
