package client

import (
	"bufio"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string
	Payload map[string]interface{}
}

var result = make(chan string)

func ReadInput() chan string {
	go func() {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		result <- text
	}()

	return result
}

type State interface {
	Execute(client *Client)
}

type IdleState struct{}

func (s *IdleState) Execute(client *Client) {
	log.Println("What to do?")

	switch <-ReadInput() {
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

	switch msg.Type {
	case "wait_for_match":
		log.Println("Wait for match...")
	case "match_found":
		matchId := int(msg.Payload["matchId"].(float64))
		client.SetState(&MatchFoundState{
			MatchId: matchId,
		})
	default:
		log.Println("weird type")
	}
}

type MatchFoundState struct {
	MatchId int
}

func (s *MatchFoundState) Execute(client *Client) {
	log.Println("accept or decline")

	select {
	case msg := <-client.Incoming:
		switch msg.Type {
		case "match_canceled":
			log.Println("Match canceled")
			client.SetState(&IdleState{})
		default:
			log.Println("nope")
		}
	case choice := <-ReadInput():
		switch choice {
		case "accept\n":
			client.Send(Message{
				Type: "match_confirmed",
				Payload: map[string]interface{}{
					"matchId": s.MatchId,
				},
			})
			client.SetState(&MatchConfirmedState{})
		case "decline\n":
			client.Send(Message{
				Type: "match_declined",
				Payload: map[string]interface{}{
					"matchId": s.MatchId,
				},
			})
			client.SetState(&IdleState{})
		default:
			log.Println("unexpected")
		}
	}
}

type MatchConfirmedState struct{}

func (s *MatchConfirmedState) Execute(client *Client) {
	msg := <-client.Incoming

	switch msg.Type {
	case "wait_for_players":
		log.Println("Waiting for players...")
	case "match_canceled":
		log.Println("Match canceled")
		client.SetState(&WaitingForMatch{})
	case "guess":
		client.SetState(&PlayingState{
			GameId: int(msg.Payload["GameId"].(float64)),
		})
	}
}

type PlayingState struct {
	GameId int
}

func (s *PlayingState) Execute(client *Client) {
	log.Println("Guess a number")

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
            log.Println(msg.Payload["message"].(string))
        case "victory":
            log.Println("Correct! You won!")
            client.SetState(&IdleState{})
        case "loss":
            answer := msg.Payload["answer"]
            log.Printf("You lost. The number was %v", answer)
            client.SetState(&IdleState{})
        }
	}
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
