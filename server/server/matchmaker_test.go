package server

import (
	"testing"
	"time"
)

func TestMatchFound(t *testing.T) {
	server := NewServer([]EventHandler{
		NewQueueManager(),
		NewMatchMaker(100 * time.Millisecond),
	})

	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(time.Millisecond)

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
	matchMaker := NewMatchMaker(100 * time.Millisecond)
	server := NewServer([]EventHandler{
		NewQueueManager(),
		NewGameManager(),
		matchMaker,
	})

	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(time.Millisecond)

	c1 := NewTestClient()
	c2 := NewTestClient()

	c1.QueueUp() // wait_for_match
	c2.QueueUp() // wait_for_match

	c1.GetIncoming() // skips match_found
	c2.GetIncoming() // skips match_found

	wait1 := c1.AcceptMatch(1)
	if wait1.Type != "wait_for_players" {
		t.Errorf("Expected \"wait_for_players\", got \"%s\"", wait1.Type)
	}

	match, _ := matchMaker.FindMatch(1)

	if match.Confirmed.Count() != 1 {
		t.Errorf("Expected 1 confirmed, got %d", match.Confirmed.Count())
	}

	wait2 := c2.AcceptMatch(1)

	if wait2.Type != "wait_for_players" {
		t.Errorf("Expected \"wait_for_players\", got \"%s\"", wait2.Type)
	}

	if match, _ := matchMaker.FindMatch(1); match != nil {
		t.Errorf("Expected match to be resolved %v", matchMaker.Count())
	}

	res1 := c1.GetIncoming()
	if res1.Type != "guess" {
		t.Errorf("Expected \"guess\", got %s", res1.Type)
	}
	res2 := c2.GetIncoming()
	if res2.Type != "guess" {
		t.Errorf("Expected \"guess\", got %s", res2.Type)
	}
}

func TestDenyMatch(t *testing.T) {
	matchMaker := NewMatchMaker(100 * time.Millisecond)
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

	c1.AcceptMatch(1)
	c2.DenyMatch(1)

	time.Sleep(time.Millisecond)

	if queueManager.Count() != 1 {
		t.Errorf("Expected confirmed to go back to queue, got %v", queueManager.Count())
	}
}

func TestTimeout(t *testing.T) {
	matchMaker := NewMatchMaker(100 * time.Millisecond)
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

	c1.QueueUp() // wait_for_match
	c2.QueueUp() // wait_for_match

	c1.AcceptMatch(1) // match found

	c1.GetIncoming() // wait_for_players
	c1.GetIncoming() // match_declined

	time.Sleep(time.Millisecond)

	res2 := c1.GetIncoming()
	if res2.Type != "wait_for_match" {
		t.Errorf("Expected \"wait_for_match\", got %v", res2.Type)
	}
}

func TestDisconnectCancelsMatch(t *testing.T) {
	matchMaker := NewMatchMaker(100 * time.Millisecond)

	server := NewServer([]EventHandler{
		NewQueueManager(),
		matchMaker,
	})
	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(time.Millisecond)

	c1 := NewTestClient()
	c2 := NewTestClient()

	c1.QueueUp() // wait for match
	c2.QueueUp() // wait for match

	c1.AcceptMatch(1) // match found
	c1.Client.Close()

	time.Sleep(time.Millisecond)

	c2.GetIncoming() // match found
	res := c2.GetIncoming()

	if res.Type != "match_canceled" {
		t.Errorf("Expected \"match_canceled\", got \"%s\"", res.Type)
	}
	if matchMaker.Count() != 0 {
		t.Errorf("Expected matches to have count 0, got %d", len(matchMaker.matches))
	}
}

func TestDisconnectRequeuesConfirmed(t *testing.T) {
	matchMaker := NewMatchMaker(100 * time.Millisecond)

	server := NewServer([]EventHandler{
		NewQueueManager(),
		matchMaker,
	})
	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(time.Millisecond)

	c1 := NewTestClient()
	c2 := NewTestClient()

	c1.QueueUp() // wait for match
	c2.QueueUp() // wait for match

	c1.AcceptMatch(1) // match found
	c2.Client.Close()

	time.Sleep(time.Millisecond)

	c1.GetIncoming() // wait_for_players
	c1.GetIncoming() // match_canceled

	res := c1.GetIncoming() // wait_for_match

	if res.Type != "wait_for_match" {
		t.Errorf("Expected \"wait_for_match\", got \"%s\"", res.Type)
	}
	if len(matchMaker.matches) != 0 {
		t.Errorf("Expected matches to have count 0, got %d", len(matchMaker.matches))
	}
}
