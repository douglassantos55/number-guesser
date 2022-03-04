package server

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Connection interface {
	Send(msg Message)
}

type Socket struct {
	conn *websocket.Conn
}

func NewSocket(conn *websocket.Conn) *Socket {
	return &Socket{conn: conn}
}

func (s *Socket) Send(msg Message) {
	s.conn.WriteJSON(msg)
}

type Sockets struct {
	conns []*Socket
	mutex *sync.Mutex
}

func NewSockets(conns []*websocket.Conn) *Sockets {
	sockets := []*Socket{}

	for _, conn := range conns {
		sockets = append(sockets, NewSocket(conn))
	}

	return &Sockets{
		conns: sockets,
		mutex: new(sync.Mutex),
	}
}

func (s *Sockets) Send(msg Message) {
	for _, socket := range s.conns {
		socket.Send(msg)
	}
}

func (s *Sockets) Add(conn *websocket.Conn) *Socket {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	socket := NewSocket(conn)
	s.conns = append(s.conns, socket)

	return socket
}

func (s *Sockets) Count() int {
	return len(s.conns)
}

type Match struct {
	Id        int
	Players   *Sockets
	Confirmed *Sockets
	Ready     chan bool
}

func NewMatch(id int, players *Sockets) *Match {
	return &Match{
		Id:        id,
		Players:   players,
		Ready:     make(chan bool),
		Confirmed: NewSockets([]*websocket.Conn{}),
	}
}

func (m *Match) AskForConfirmation() {
	m.Players.Send(Message{
		Type: "match_found",
		Payload: map[string]interface{}{
			"matchId": m.Id,
		},
	})
}

func (m *Match) WaitForConfirmation(timeout time.Duration, dispatch func(event Event)) {
	select {
	case isReady := <-m.Ready:
		if isReady {
			dispatch(Event{
				Type: "game_start",
				Payload: map[string]interface{}{
					"players": m.Confirmed,
				},
			})
		}
	case <-time.After(timeout):
		m.Cancel(dispatch)
	}
}

func (m *Match) Cancel(dispatch func(event Event)) {
	m.Players.Send(Message{
		Type: "match_canceled",
		Payload: map[string]interface{}{
			"match": m.Id,
		},
	})

	m.RequeueConfirmed(dispatch)
}

func (m *Match) RequeueConfirmed(dispatch func(event Event)) {
	for _, socket := range m.Confirmed.conns {
		dispatch(Event{
			Type:   "queue_up",
			Socket: socket.conn,
		})
	}
}

type MatchMaker struct {
	currentId int
	timeout   time.Duration
	matches   map[int]*Match
	mut       *sync.Mutex
}

func NewMatchMaker(timeout time.Duration) *MatchMaker {
	return &MatchMaker{
		currentId: 0,
		timeout:   timeout,
		mut:       new(sync.Mutex),
		matches:   make(map[int]*Match),
	}
}

func (m *MatchMaker) AddMatch(match *Match) {
	m.matches[match.Id] = match
}

func (m *MatchMaker) RemoveMatch(match *Match) {
	delete(m.matches, match.Id)
}

func (m *MatchMaker) FindMatch(matchId int) (*Match, error) {
	match, ok := m.matches[matchId]

	if !ok {
		return nil, errors.New(fmt.Sprintf("Match with ID %d not found", matchId))
	}

	return match, nil
}

func (m *MatchMaker) Process(event Event, server *Server) {
	m.mut.Lock()
	defer m.mut.Unlock()

	if event.Type == "match_found" {
		m.currentId = m.currentId + 1
		players := event.Payload["players"].([]*websocket.Conn)

		match := NewMatch(m.currentId, NewSockets(players))
		m.AddMatch(match)

		match.AskForConfirmation()
		go match.WaitForConfirmation(m.timeout, server.Dispatch)
	}

	if event.Type == "match_confirmed" {
		matchId := int(event.Payload["matchId"].(float64))
		match, err := m.FindMatch(matchId)

		if err == nil {
			player := match.Confirmed.Add(event.Socket)

			player.Send(Message{
				Type: "wait_for_players",
			})

			if match.Confirmed.Count() == NUM_OF_PLAYERS {
				match.Ready <- true
				m.RemoveMatch(match)
			}
		}
	}

	if event.Type == "match_declined" {
		matchId := int(event.Payload["matchId"].(float64))
		match, err := m.FindMatch(matchId)

		if err == nil {
			match.Players.Send(Message{
				Type: "match_canceled",
				Payload: map[string]interface{}{
					"matchId": match.Id,
				},
			})

			match.Cancel(server.Dispatch)
			m.RemoveMatch(match)
		}
	}
}
