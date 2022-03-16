package server

import (
	"sync"
)

const NUM_OF_PLAYERS = 2

type Queue struct {
	Head *Node
	Tail *Node

	mut     *sync.Mutex
	sockets map[*Socket]*Node
}

type Node struct {
	Next   *Node
	Prev   *Node
	Socket *Socket
}

func (q *Queue) Count() int {
	q.mut.Lock()
	defer q.mut.Unlock()

	return len(q.sockets)
}

func (q *Queue) Pop() *Socket {
	q.mut.Lock()

	if q.Head != nil {
		socket := q.Head.Socket
		q.mut.Unlock()
		q.Remove(socket)

		return socket
	}

	return nil
}

func (q *Queue) Remove(socket *Socket) {
	q.mut.Lock()
	defer q.mut.Unlock()

	node := q.sockets[socket]

	if node != nil {
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
}

func (q *Queue) Push(socket *Socket) {
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
	mutex *sync.Mutex
}

func NewQueueManager() *QueueManager {
	return &QueueManager{
		mutex: new(sync.Mutex),
		queue: &Queue{
			mut:     new(sync.Mutex),
			sockets: make(map[*Socket]*Node),
		},
	}
}

func (q *QueueManager) Push(socket *Socket) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.queue.Push(socket)
}

func (q *QueueManager) Pop() *Socket {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.queue.Pop()
}

func (q *QueueManager) Remove(socket *Socket) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.queue.Remove(socket)
}

func (q *QueueManager) Count() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.queue.Count()
}

func (q *QueueManager) Process(event Event, server *Server) {
	switch event.Type {
	case "dequeue", "disconnected":
		q.Remove(event.Socket)
	case "queue_up":
		q.Push(event.Socket)

		event.Socket.Send(Message{
			Type: "wait_for_match",
		})

		if q.Count() == NUM_OF_PLAYERS {
			players := make([]*Socket, 0)

			for i := 0; i < NUM_OF_PLAYERS; i++ {
				players = append(players, q.Pop())
			}

			server.Dispatch(Event{
				Type:   "match_found",
				Socket: event.Socket,
				Payload: map[string]interface{}{
					"players": players,
				},
			})
		}
	}
}
