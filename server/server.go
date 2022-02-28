package server

import (
	"context"
	"net/http"

	"example.com/game/common"
	"github.com/gorilla/websocket"
)

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
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	go func() {
		defer c.Close()

		for {
			var msg common.Message
			err := c.ReadJSON(&msg)

			if err != nil {
				continue
			}

			s.Dispatch(NewEvent(msg, c))
		}
	}()
}

func (s *Server) Dispatch(event Event) {
	for _, handler := range s.handlers {
		handler.Process(event, s)
	}
}
