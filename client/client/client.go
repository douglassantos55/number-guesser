package client

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string
	Payload map[string]interface{}
}

func ReadInput() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	return text
}

type State interface {
	Execute(client *Client)
}

type IdleState struct{}

func (s *IdleState) Execute(client *Client) {
	log.Println("What to do?")

	switch ReadInput() {
	case "play\n":
		client.Send(Message{
            Type: "queue_up",
        })
		client.SetState(&WaitingForMatch{})
	default:
		log.Println("Invalid option")
	}
}

type WaitingForMatch struct{}

func (s *WaitingForMatch) Execute(client *Client) {
	msg := <-client.Incoming
    fmt.Printf("msg: %v\n", msg)

	switch msg.Type {
	case "wait_for_match":
		log.Println("Wait for match...")
	case "match_found":
		log.Println("Match found")
		client.SetState(&MatchFoundState{})
	default:
		client.SetState(&IdleState{})
	}
}

type MatchFoundState struct{}

func (s *MatchFoundState) Execute(client *Client) {
	msg := <-client.Incoming
    fmt.Printf("msg: %v\n", msg)
	switch msg.Type {
	case "guess":
		gameId := int(msg.Payload["GameId"].(float64))

		client.SetState(&PlayingState{
			GameId: gameId,
		})
	case "match_canceled":
		log.Println("Match canceled")
		client.SetState(&WaitingForMatch{})
	}
}

type PlayingState struct {
	GameId int
}

func (s *PlayingState) Execute(client *Client) {
	log.Println("Guess a number")

	guess := ReadInput()

	client.Send(Message{
		Type: "guess",
		Payload: map[string]interface{}{
			"guess":  guess,
			"gameId": s.GameId,
		},
	})

	msg := <-client.Incoming
	log.Println(msg)
}

type Client struct {
	state    State
	Outgoing chan Message
	Incoming chan Message
}

func NewClient() *Client {
	return &Client{
		state:    &IdleState{},
		Outgoing: make(chan Message),
		Incoming: make(chan Message),
	}
}

func (c *Client) SetState(state State) {
	c.state = state
}

func (c *Client) Loop() {
	for {
		c.state.Execute(c)
	}
}

func (c *Client) Connect(addr string) error {
	socket, _, err := websocket.DefaultDialer.Dial("ws://"+addr, nil)

	if err != nil {
		return err
	}

	go func() {
		defer socket.Close()

		go func() {
			for {
				var response Message
				err := socket.ReadJSON(&response)

				if err != nil {
					continue
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
	c.Outgoing <- message
}
