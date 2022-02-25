package server

import (
	"github.com/gorilla/websocket"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(addr string) error {
	_, _, err := websocket.DefaultDialer.Dial("ws://"+addr, nil)

	if err != nil {
		return err
	}

	return nil
}
