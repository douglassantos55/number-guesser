package server

import (
	"github.com/gorilla/websocket"
)

type Event struct {
	Type    string
	Payload map[string]interface{}
	Socket  *websocket.Conn
}

type EventHandler interface {
	Process(event Event, server *Server)
}

func NewEvent(msg Message, socket *websocket.Conn) Event {
	return Event{
		Type:    msg.Type,
		Payload: msg.Payload,
		Socket:  socket,
	}
}
