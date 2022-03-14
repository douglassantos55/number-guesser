module example.com/game/server

go 1.17

replace example.com/game/client => /home/douglas/Desktop/game-server/client

require github.com/google/uuid v1.3.0 // indirect

require (
	example.com/game/client v1.0.0
	github.com/gorilla/websocket v1.5.0
)
