package server

import (
	"sync"

	"example.com/game/client/client"
	"github.com/gorilla/websocket"
)

const NUM_OF_PLAYERS = 2

type Queue struct {
	Head *Node
	Tail *Node

	sockets map[*websocket.Conn]*Node
}

type Node struct {
	Next   *Node
	Prev   *Node
	Socket *websocket.Conn
}

func (q *Queue) Count() int {
	return len(q.sockets)
}

func (q *Queue) Pop() *websocket.Conn {
	node := q.Head
	q.Head = q.Head.Next

	if q.Head != nil {
		q.Head.Prev = nil
	}

	q.Remove(node.Socket)
	return node.Socket
}

func (q *Queue) Remove(socket *websocket.Conn) {
	node := q.sockets[socket]

	if node.Prev != nil {
		node.Prev.Next = node.Next
	}

	delete(q.sockets, socket)
}

func (q *Queue) Push(socket *websocket.Conn) {
	node := &Node{
		Socket: socket,
	}

	if q.Head == nil {
		q.Head = node
		q.Tail = node
	} else {
		q.Tail.Next = node
		node.Prev = q.Tail
		q.Tail = node
	}

	q.sockets[socket] = node
}

type QueueManager struct {
	queue *Queue
	mut   *sync.Mutex
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		mut: new(sync.Mutex),
		queue: &Queue{
			sockets: make(map[*websocket.Conn]*Node),
		},
	}
}

func (q *QueueManager) Process(event Event, server *Server) {
	if event.Type == "dequeue" {
		q.mut.Lock()
		defer q.mut.Unlock()

		q.queue.Remove(event.Socket)
	}

	if event.Type == "queue_up" {
		q.mut.Lock()
		defer q.mut.Unlock()

		q.queue.Push(event.Socket)

		event.Socket.WriteJSON(client.Message{
			Type: "wait_for_match",
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
