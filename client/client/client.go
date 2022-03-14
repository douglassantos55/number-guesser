package client

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string
	Payload map[string]interface{}
}

var result = make(chan string)

func ReadInput() chan string {
	go func() {
		// some goroutines are going rogue when there is a timeout
		// since I couldn't figure out how to cancel the read
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			result <- scanner.Text()
		}
	}()

	return result
}

type State interface {
	Execute(client *Client)
}

type IdleState struct{}

func (s *IdleState) Execute(client *Client) {
	fmt.Println("Type \"play\" or \"quit\"")

	switch <-ReadInput() {
	case "play":
		client.Send(Message{
			Type: "queue_up",
		})
		client.SetState(&WaitingForMatch{})
	case "quit":
		client.Close()
	default:
		fmt.Println("Invalid option")
	}
}

type WaitingForMatch struct{}

func (s *WaitingForMatch) Execute(client *Client) {
	select {
	case choice := <-ReadInput():
		switch choice {
		case "cancel":
			client.Send(Message{
				Type: "dequeue",
			})
			client.SetState(&IdleState{})
		}
	case msg := <-client.Incoming:
		switch msg.Type {
		case "wait_for_match":
			fmt.Println("Waiting for match... type \"cancel\" to leave")
		case "match_found":
			matchId := int(msg.Payload["matchId"].(float64))
			client.SetState(&MatchFoundState{
				MatchId: matchId,
			})
		default:
			fmt.Println("weird type", msg)
		}
	}
}

type MatchFoundState struct {
	MatchId int
}

func (s *MatchFoundState) Execute(client *Client) {
	fmt.Println("Match found. Type \"accept\" or \"decline\"")

	select {
	case msg := <-client.Incoming:
		switch msg.Type {
		case "match_canceled":
			fmt.Println("Match canceled")
			client.SetState(&IdleState{})
		default:
			fmt.Println("nope", msg)
		}
	case choice := <-ReadInput():
		switch choice {
		case "accept":
			client.Send(Message{
				Type: "match_confirmed",
				Payload: map[string]interface{}{
					"matchId": s.MatchId,
				},
			})
			client.SetState(&MatchConfirmedState{})
		case "decline":
			client.Send(Message{
				Type: "match_declined",
				Payload: map[string]interface{}{
					"matchId": s.MatchId,
				},
			})
			client.SetState(&IdleState{})
		default:
			fmt.Println("unexpected", choice)
		}
	}
}

type MatchConfirmedState struct{}

func (s *MatchConfirmedState) Execute(client *Client) {
	msg := <-client.Incoming

	switch msg.Type {
	case "wait_for_players":
		fmt.Println("Waiting for players...")
	case "match_canceled":
		fmt.Println("Match canceled")
		client.SetState(&WaitingForMatch{})
	case "guess":
		fmt.Println("Guess a number")
		client.SetState(&PlayingState{
			GameId: int(msg.Payload["GameId"].(float64)),
		})
	}
}

type PlayingState struct {
	GameId int
}

func (s *PlayingState) Execute(client *Client) {
	select {
	case guess := <-ReadInput():
		client.Send(Message{
			Type: "guess",
			Payload: map[string]interface{}{
				"guess":  guess,
				"gameId": s.GameId,
			},
		})
	case msg := <-client.Incoming:
		switch msg.Type {
		case "feedback":
			fmt.Println(msg.Payload["message"].(string))
		case "victory":
			fmt.Println(msg.Payload["message"].(string))
			client.SetState(&IdleState{})
		case "loss":
			fmt.Println(msg.Payload["message"].(string))
			client.SetState(&IdleState{})
		}
	}
}

type Client struct {
	state  State
	mutex  *sync.Mutex
	socket *websocket.Conn

	Id       uuid.UUID
	Running  bool
	Outgoing chan Message
	Incoming chan Message
}

func NewClient() *Client {
	return &Client{
		mutex: new(sync.Mutex),
		state: &IdleState{},

		Id:       uuid.New(),
		Running:  true,
		Outgoing: make(chan Message),
		Incoming: make(chan Message),
	}
}

func (c *Client) SetState(state State) {
	c.state = state
}

func (c *Client) Loop() {
	for c.Running {
		c.state.Execute(c)
	}
}

func (c *Client) Connect(addr string) error {
	socket, _, err := websocket.DefaultDialer.Dial("ws://"+addr, nil)

	if err != nil {
		return err
	}

	c.socket = socket

	go func() {
		defer socket.Close()

		go func() {
			for {
				var response Message
				err := socket.ReadJSON(&response)

				if err != nil {
					c.Close()
					break
				}

				c.Incoming <- response
			}
		}()

		for {
			select {
			case msg := <-c.Outgoing:
				socket.WriteJSON(msg)
			}
		}

	}()

	return nil
}

func (c *Client) Send(message Message) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Outgoing <- message
}

func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.socket.Close()
	c.Running = false
}
