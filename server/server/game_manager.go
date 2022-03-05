package server

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
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

func (g *GameManager) AddGame(game *Game) {
	g.mut.Lock()
	defer g.mut.Unlock()

	g.Games[game.Id] = game
}

func (g *GameManager) RemoveGame(game *Game) {
	g.mut.Lock()
	defer g.mut.Unlock()

	delete(g.Games, game.Id)
}

func (g *GameManager) FindGame(id int) (*Game, error) {
	game, ok := g.Games[id]

	if !ok {
		return nil, errors.New(fmt.Sprintf("Game with ID %d not found", id))
	}
	return game, nil
}

func (g *GameManager) Process(event Event, server *Server) {
	if event.Type == "game_start" {
		// create a game instance
		players := event.Payload["players"].(*Sockets)
		game := NewGame(players)

        log.Println("answer", game.Answer)
		g.AddGame(game)
		game.Start()
	}

	if event.Type == "guess" {
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
