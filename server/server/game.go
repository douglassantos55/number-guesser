package server

import (
	"fmt"
	"math/rand"
	"sync"
)

type Game struct {
	mutex *sync.Mutex

	Id      int
	Answer  int
	Done    bool
	Players *Sockets
}

func NewGame(players *Sockets) *Game {
	return &Game{
		mutex: new(sync.Mutex),

		Id:      rand.Intn(1000),
		Done:    false,
		Players: players,
		Answer:  rand.Intn(100),
	}
}

func (g *Game) Start() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// send guess for both players
	g.Players.Send(Message{
		Type: "guess",
		Payload: map[string]interface{}{
			"GameId": g.Id,
		},
	})
}

func (g *Game) CheckGuess(guess int, player *Socket) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// if guess = Answer, end game
	if guess == g.Answer {
		g.End(player)
		return true
	} else {
		// otherwise, respond with > or <
		g.Feedback(guess, player)
		return false
	}
}

func (g *Game) End(winner *Socket) {
	for _, player := range g.Players.conns {
		if player.conn != winner.conn {
			// send loss to loser
			player.Send(Message{
				Type: "loss",
				Payload: map[string]interface{}{
					"message": fmt.Sprintf("You lost. The number was %d", g.Answer),
				},
			})
		} else {
			// send victory to winner
			winner.Send(Message{
				Type: "victory",
				Payload: map[string]interface{}{
					"message": "Correct! You won!",
				},
			})
		}
	}
}

func (g *Game) Feedback(guess int, player *Socket) {
	feedback := ""

	if guess < g.Answer {
		feedback = "Try a greater number"
	} else {
		feedback = "Try a smaller number"
	}

	player.Send(Message{
		Type: "feedback",
		Payload: map[string]interface{}{
			"message": feedback,
		},
	})
}
