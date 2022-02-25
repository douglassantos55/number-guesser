package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
    server *http.Server
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Listen(addr string) {
    s.server = &http.Server{Addr: addr, Handler: http.HandlerFunc(s.HandleRequest)}
    s.server.ListenAndServe()
}

func (s *Server) Close() {
    s.server.Close()
}

func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
        return
	}

	defer c.Close()

	go func() {
		for {
		}
	}()
}
