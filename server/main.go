package main

import (
	"time"

	"example.com/game/server/server"
)

func main() {
    server := server.NewServer([]server.EventHandler{
        server.NewGameManager(),
        server.NewQueueManager(),
        server.NewMatchMaker(10 * time.Second),
    })

    defer server.Close()
    server.Listen("0.0.0.0:8080")
}
