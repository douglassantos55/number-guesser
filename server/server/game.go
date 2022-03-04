package server

import (
	"math/rand"
)

type Game struct {
	Id      int
	Answer  int
	Done    bool
	Players *Sockets
}

func NewGame(players *Sockets) *Game {
	return &Game{
		Id:      rand.Intn(1000),
		Done:    false,
		Players: players,
		Answer:  rand.Intn(100),
	}
}

func (g *Game) Start() {
	// send guess for both players
	g.Players.Send(Message{
		Type: "guess",
		Payload: map[string]interface{}{
			"GameId": g.Id,
		},
	})
}

func (g *Game) CheckGuess(guess int, player *Socket) bool {
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
	// send victory to winner
	winner.Send(Message{
		Type: "victory",
	})

	for _, player := range g.Players.conns {
		if player != winner {
			// send loss to loser
			player.Send(Message{
				Type: "loss",
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
