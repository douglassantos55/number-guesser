package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type GameManager struct {
	Games map[int]*Game
	mut   *sync.Mutex
}

func NewGameManager() *GameManager {
	return &GameManager{
		Games: make(map[int]*Game),
		mut:   new(sync.Mutex),
	}
}

func (g *GameManager) AddGame(players *Sockets) *Game {
	g.mut.Lock()
	defer g.mut.Unlock()

	game := NewGame(players)
	g.Games[game.Id] = game

	return game
}

func (g *GameManager) RemoveGame(game *Game) {
	g.mut.Lock()
	defer g.mut.Unlock()

	game.Players.Send(Message{
		Type: "victory",
		Payload: map[string]interface{}{
			"message": "You won. The other player disconnected.",
		},
	})

	delete(g.Games, game.Id)
}

func (g *GameManager) FindGame(id int) (*Game, error) {
	g.mut.Lock()
	defer g.mut.Unlock()

	game, ok := g.Games[id]

	if !ok {
		return nil, errors.New(fmt.Sprintf("Game with ID %d not found", id))
	}
	return game, nil
}

func (g *GameManager) FindGameWithSocket(socket *websocket.Conn) *Game {
	g.mut.Lock()
	defer g.mut.Unlock()

	for _, game := range g.Games {
		if game.Players.Has(socket) {
			return game
		}
	}

	return nil
}

func (g *GameManager) Process(event Event, server *Server) {
	switch event.Type {
	case "disconnected":
		game := g.FindGameWithSocket(event.Socket)

		if game != nil {
			g.RemoveGame(game)
		}

	case "game_start":
		players := event.Payload["players"].(*Sockets)
		game := g.AddGame(players)
		game.Start()

	case "guess":
		// get game
		gameId := int(event.Payload["gameId"].(float64))

		game, err := g.FindGame(gameId)

		if err == nil {
			guess, _ := strconv.Atoi(strings.TrimSpace(event.Payload["guess"].(string)))

			if game.CheckGuess(guess, NewSocket(event.Socket)) {
				g.RemoveGame(game)
			}
		}
	}
}
