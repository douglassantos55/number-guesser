package server

type Event struct {
	Type    string
	Payload map[string]interface{}
	Socket  *Socket
}

type EventHandler interface {
	Process(event Event, server *Server)
}

func NewEvent(msg Message, socket *Socket) Event {
	return Event{
		Type:    msg.Type,
		Payload: msg.Payload,
		Socket:  socket,
	}
}
