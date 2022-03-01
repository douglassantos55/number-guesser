package server

import (
	"testing"
	"time"
)

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
	if res1.Type != "match_found" {
		t.Errorf("Expected match_found, got %s", res1.Type)
	}

	if int(res1.Payload["matchId"].(float64)) != 1 {
		t.Errorf("Expected matchId: 1, got %d", int(res1.Payload["matchId"].(float64)))
	}

	res2 := c2.GetIncoming()
	if res2.Type != "match_found" {
		t.Errorf("Expected match_found response, got %s", res2.Type)
	}

	if int(res2.Payload["matchId"].(float64)) != 1 {
		t.Errorf("Expected matchId: 1, got %d", int(res2.Payload["matchId"].(float64)))
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
	if res1.Type != "wait_for_players" {
		t.Errorf("Expected \"wait_for_players\", got %s", res1.Type)
	}
	res2 := c2.GetIncoming()
	if res2.Type != "wait_for_players" {
		t.Errorf("Expected \"wait_for_players\", got %s", res2.Type)
	}
}

func TestDenyMatch(t *testing.T) {
	matchMaker := NewMatchMaker()
	queueManager := NewQueueManager()

	server := NewServer([]EventHandler{
		queueManager,
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
	c2.DenyMatch(1)

	if queueManager.queue.Count() != 1 {
		t.Error("Expected confirmed to go back to queue")
	}
}

func TestTimeout(t *testing.T) {
	matchMaker := NewMatchMaker()
	queueManager := NewQueueManager()

	server := NewServer([]EventHandler{
		queueManager,
		matchMaker,
	})

	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(time.Millisecond)

	c1 := NewTestClient()
	c2 := NewTestClient()

	c1.QueueUp()
	c2.QueueUp()

	c1.AcceptMatch(1) // match found

    c1.GetIncoming() // wait_for_players
	c1.GetIncoming() // match_declined
	res2 := c1.GetIncoming()

	if res2.Type != "wait_for_match" {
		t.Errorf("Expected \"wait_for_match\", got %v", res2.Type)
	}
}
