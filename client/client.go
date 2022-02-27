package client

import (
	"example.com/game/common"
	"github.com/gorilla/websocket"
)

type Client struct {
	Outgoing chan common.Message
	Incoming chan common.Response
}

func NewClient() *Client {
	return &Client{
		Outgoing: make(chan common.Message),
		Incoming: make(chan common.Response),
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
				var response common.Response
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

func (c *Client) Send(message common.Message) {
	c.Outgoing <- message
}
