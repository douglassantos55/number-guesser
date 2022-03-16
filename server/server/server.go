package server

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string
	Payload map[string]interface{}
}

type Server struct {
	server   *http.Server
	handlers []EventHandler
}

func NewServer(handlers []EventHandler) *Server {
	return &Server{
		handlers: handlers,
	}
}

func (s *Server) Listen(addr string) {
	s.server = &http.Server{Addr: addr, Handler: http.HandlerFunc(s.HandleRequest)}
	s.server.ListenAndServe()
}

func (s *Server) Close() {
	s.server.Shutdown(context.Background())
}

func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	connection, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	socket := NewSocket(connection)

	go func() {
		defer socket.Close()

		for {
			var msg Message
			err := connection.ReadJSON(&msg)

			if err != nil {
				s.Dispatch(Event{
					Socket: socket,
					Type:   "disconnected",
				})
				break
			}

			s.Dispatch(NewEvent(msg, socket))
		}
	}()
}

func (s *Server) Dispatch(event Event) {
	for _, handler := range s.handlers {
		handler.Process(event, s)
	}
}
