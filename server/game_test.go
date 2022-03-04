package server

import (
	"testing"
	"time"
)

func TestGame(t *testing.T) {
	server := NewServer([]EventHandler{
		NewQueueManager(),
		NewGameManager(),
		NewMatchMaker(200 * time.Millisecond),
	})

	defer server.Close()
	go server.Listen("0.0.0.0:8080")

	time.Sleep(100 * time.Millisecond)

	c1 := NewTestClient()
	c2 := NewTestClient()

	c1.QueueUp() // wait_for_match
	c2.QueueUp() // wait_for_match

	c1.GetIncoming() // skips match_found
	c2.GetIncoming() // skips match_found

	c1.AcceptMatch(1) // wait_for_players
	c2.AcceptMatch(1) // wait_for_players

	res := c1.GetIncoming() // guess
	c2.GetIncoming()        // guess

	gameId := int(res.Payload["GameId"].(float64))

	res1 := c1.Guess(69, gameId)
	res2 := c2.Guess(4, gameId)

	if res1.Type != "feedback" {
		t.Errorf("Expected \"feedback\", got \"%v\"", res1.Type)
	}
	if res1.Payload["message"] != "Try a smaller number" {
		t.Errorf("Expected \"Try a smaller number\", got \"%v\"", res1.Payload["message"])
	}

	if res2.Type != "feedback" {
		t.Errorf("Expected \"feedback\", got \"%v\"", res2.Type)
	}
	if res2.Payload["message"] != "Try a greater number" {
		t.Errorf("Expected \"Try a greater number\", got \"%v\"", res2.Payload["message"])
	}

	victory := c1.Guess(40, gameId)
	defeat := c2.Guess(355, gameId)

	if victory.Type != "victory" {
		t.Errorf("Expected \"victory\", got \"%v\"", victory.Type)
	}

	if defeat.Type != "loss" {
		t.Errorf("Expected \"loss\", got \"%v\"", defeat.Type)
	}
}
