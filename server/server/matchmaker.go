package server

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Socket struct {
	conn  *websocket.Conn
	mutex *sync.Mutex
}

func NewSocket(conn *websocket.Conn) *Socket {
	return &Socket{
		conn:  conn,
		mutex: new(sync.Mutex),
	}
}

func (s *Socket) Send(msg Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.conn.WriteJSON(msg)
}

func (s *Socket) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.conn.Close()
}

type Sockets struct {
	conns []*Socket
	mutex *sync.Mutex
}

func NewSockets(sockets []*Socket) *Sockets {
	return &Sockets{
		conns: sockets,
		mutex: new(sync.Mutex),
	}
}

func (s *Sockets) Send(msg Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, socket := range s.conns {
		socket.Send(msg)
	}
}

func (s *Sockets) Add(socket *Socket) *Socket {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.conns = append(s.conns, socket)
	return socket
}

func (s *Sockets) Count() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.conns)
}

func (s *Sockets) Has(conn *Socket) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, socket := range s.conns {
		if socket.conn == conn.conn {
			return true
		}
	}
	return false
}

type Match struct {
	mutex *sync.Mutex

	Id        int
	Players   *Sockets
	Confirmed *Sockets
	Ready     chan bool
}

func NewMatch(id int, players *Sockets) *Match {
	return &Match{
		mutex: new(sync.Mutex),

		Id:        id,
		Players:   players,
		Ready:     make(chan bool),
		Confirmed: NewSockets([]*Socket{}),
	}
}

func (m *Match) AskForConfirmation() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

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
	m.mutex.Lock()
	defer m.mutex.Unlock()

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
			Socket: socket,
		})
	}
}

func (m *Match) AddConfirmed(socket *Socket) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Confirmed.Add(socket)

	socket.Send(Message{
		Type: "wait_for_players",
	})
}

func (m *Match) CountConfirmed() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.Confirmed.Count()
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

func (m *MatchMaker) AddMatch(players *Sockets) *Match {
	m.mut.Lock()
	defer m.mut.Unlock()

	m.currentId = m.currentId + 1
	match := NewMatch(m.currentId, players)
	m.matches[m.currentId] = match

	return match
}

func (m *MatchMaker) RemoveMatch(match *Match) {
	m.mut.Lock()
	defer m.mut.Unlock()

	delete(m.matches, match.Id)
}

func (m *MatchMaker) FindMatch(matchId int) (*Match, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	match, ok := m.matches[matchId]

	if !ok {
		return nil, errors.New(fmt.Sprintf("Match with ID %d not found", matchId))
	}

	return match, nil
}

func (m *MatchMaker) FindMatchWithSocket(socket *Socket) *Match {
	m.mut.Lock()
	defer m.mut.Unlock()

	for _, match := range m.matches {
		if match.Players.Has(socket) {
			return match
		}
	}
	return nil
}

// Returns the current number of matches pending
func (m *MatchMaker) Count() int {
	m.mut.Lock()
	defer m.mut.Unlock()

	return len(m.matches)
}

func (m *MatchMaker) Process(event Event, server *Server) {
	switch event.Type {
	case "disconnected":
		match := m.FindMatchWithSocket(event.Socket)

		if match != nil {
			m.RemoveMatch(match)
			match.Cancel(server.Dispatch)
			match.Ready <- false
		}

	case "match_found":
		players := event.Payload["players"].([]*Socket)
		match := m.AddMatch(NewSockets(players))

		match.AskForConfirmation()
		go match.WaitForConfirmation(m.timeout, server.Dispatch)

	case "match_confirmed":
		matchId := int(event.Payload["matchId"].(float64))
		match, err := m.FindMatch(matchId)

		if err == nil {
			match.AddConfirmed(event.Socket)

			if match.CountConfirmed() == NUM_OF_PLAYERS {
				match.Ready <- true
				m.RemoveMatch(match)
			}
		}

	case "match_declined":
		matchId := int(event.Payload["matchId"].(float64))
		match, err := m.FindMatch(matchId)

		if err == nil {
			m.RemoveMatch(match)
			match.Cancel(server.Dispatch)
			match.Ready <- false
		}
	}
}
