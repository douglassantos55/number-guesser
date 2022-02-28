package server

import (
	"sync"

	"example.com/game/common"
	"github.com/gorilla/websocket"
)

type Match struct {
	Id        int
	Players   []*websocket.Conn
	Confirmed []*websocket.Conn
}

func NewMatch(id int, players []*websocket.Conn) *Match {
	return &Match{
		Id:        id,
		Players:   players,
		Confirmed: []*websocket.Conn{},
	}
}

func (m *Match) SendConfirmation() {
	for _, socket := range m.Players {
		socket.WriteJSON(common.Message{
			Type: "confirm_match",
			Payload: map[string]interface{}{
				"matchId": m.Id,
			},
		})
	}
}

type MatchMaker struct {
	currentId int
	matches   map[int]*Match
	mut       *sync.Mutex
}

func NewMatchMaker() *MatchMaker {
	return &MatchMaker{
		currentId: 0,
		mut:       new(sync.Mutex),
		matches:   make(map[int]*Match),
	}
}

func (m *MatchMaker) Confirmed(match int) int {
	m.mut.Lock()
	defer m.mut.Unlock()

	return len(m.matches[match].Confirmed)
}

func (m *MatchMaker) HasMatch(match int) bool {
	_, ok := m.matches[match]
	return ok
}

func (m *MatchMaker) Process(event Event, server *Server) {
	m.mut.Lock()
	defer m.mut.Unlock()

	if event.Type == "match_found" {
		matchId := m.currentId + 1
		m.currentId = matchId

		players := event.Payload["players"].([]*websocket.Conn)
		match := NewMatch(matchId, players)

		m.matches[matchId] = match
		match.SendConfirmation()
	}

	if event.Type == "match_confirmed" {
		matchId := int(event.Payload["matchId"].(float64))

		if m.HasMatch(matchId) {
			match := m.matches[matchId]
			match.Confirmed = append(match.Confirmed, event.Socket)

			if len(match.Confirmed) == len(match.Players) {
				for _, socket := range match.Confirmed {
					socket.WriteJSON(common.Message{
						Type: "game_start",
					})
				}
				delete(m.matches, match.Id)
			}
		}
	}
}
