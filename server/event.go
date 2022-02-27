package server

import (
	"example.com/game/common"
	"github.com/gorilla/websocket"
)

type Event struct {
	Type    string
	Payload map[string]interface{}
	Socket  *websocket.Conn
}

type EventHandler interface {
	Process(event Event)
}

type QueueManager struct {
}

func NewEvent(msg common.Message, socket *websocket.Conn) Event {
	return Event{
		Type:    msg.Type,
		Payload: msg.Payload,
		Socket:  socket,
	}
}

func (q *QueueManager) Process(event Event) {
	if event.Type == "queue_up" {
		event.Socket.WriteJSON(common.Message{
			Type: "wait",
		})
	}
}
