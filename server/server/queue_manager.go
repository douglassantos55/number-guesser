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

	mut     *sync.Mutex
	sockets map[*websocket.Conn]*Node
}

type Node struct {
	Next   *Node
	Prev   *Node
	Socket *websocket.Conn
}

func (q *Queue) Count() int {
	q.mut.Lock()
	defer q.mut.Unlock()

	return len(q.sockets)
}

func (q *Queue) Pop() *websocket.Conn {
    q.mut.Lock()

	if q.Head != nil {
		socket := q.Head.Socket
        q.mut.Unlock()
		q.Remove(socket)

		return socket
	}

	return nil
}

func (q *Queue) Remove(socket *websocket.Conn) {
	q.mut.Lock()
	defer q.mut.Unlock()

	node := q.sockets[socket]

	if node.Prev != nil {
		node.Prev.Next = node.Next

		if node.Next == nil {
			q.Tail = node.Prev
		} else {
			node.Next.Prev = node.Prev
		}
	} else if node.Next != nil {
		node.Next.Prev = nil
		q.Head = node.Next
	} else {
		q.Head = nil
		q.Tail = nil
	}

	delete(q.sockets, socket)
}

func (q *Queue) Push(socket *websocket.Conn) {
	q.mut.Lock()
	defer q.mut.Unlock()

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
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		queue: &Queue{
			mut:     new(sync.Mutex),
			sockets: make(map[*websocket.Conn]*Node),
		},
	}
}

func (q *QueueManager) Process(event Event, server *Server) {
	if event.Type == "dequeue" {
		q.queue.Remove(event.Socket)
	}

	if event.Type == "queue_up" {
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
