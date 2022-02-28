package server

import (
	"sync"

	"example.com/game/common"
	"github.com/gorilla/websocket"
)

const NUM_OF_PLAYERS = 2

type Queue struct {
	Head *Node
}

type Node struct {
	Next   *Node
	Socket *websocket.Conn
}

func (q *Queue) Count() int {
	count := 0
	for cur := q.Head; cur != nil; cur = cur.Next {
		count++
	}
	return count
}

func (q *Queue) Pop() *websocket.Conn {
	node := q.Head
	q.Head = q.Head.Next
	return node.Socket
}

func (q *Queue) Push(socket *websocket.Conn) {
	node := &Node{
		Socket: socket,
	}

	if q.Head == nil {
		q.Head = node
	} else {
		cur := q.Head

		for cur != nil && cur.Next != nil {
			cur = cur.Next
		}

		cur.Next = node
	}
}

type QueueManager struct {
	queue *Queue
	mut   *sync.Mutex
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		queue: &Queue{},
		mut:   new(sync.Mutex),
	}
}

func (q *QueueManager) Process(event Event, server *Server) {
	if event.Type == "queue_up" {
        q.mut.Lock()
        defer q.mut.Unlock()

		q.queue.Push(event.Socket)

		event.Socket.WriteJSON(common.Message{
			Type: "wait",
		})

		if q.queue.Count() == NUM_OF_PLAYERS {
			players := make([]*websocket.Conn, 0)

			for i := 0; i < NUM_OF_PLAYERS; i++ {
				players = append(players, q.queue.Pop())
			}

			server.Dispatch(Event{
				Type: "match_found",
				Payload: map[string]interface{}{
					"players": players,
				},
			})
		}
	}
}
