package server

import (
	"log"
	"time"

	"example.com/game/client"
	"example.com/game/common"
)

type TestClient struct {
	Client *client.Client
}

func NewTestClient() *TestClient {
	testClient := &TestClient{
		Client: client.NewClient(),
	}
	testClient.Client.Connect("0.0.0.0:8080")
	return testClient
}

func (c *TestClient) GetIncoming() common.Message {
	select {
	case msg := <-c.Client.Incoming:
		return msg
	case <-time.After(time.Second):
		log.Fatal("Timeout")
	}
	// should never reach this
	return common.Message{}
}

func (c *TestClient) QueueUp() common.Message {
	c.Client.Send(common.Message{
		Type: "queue_up",
	})
	return c.GetIncoming()
}

func (c *TestClient) AcceptMatch(match int) common.Message {
	c.Client.Send(common.Message{
		Type: "match_confirmed",
		Payload: map[string]interface{}{
			"matchId": match,
		},
	})
	time.Sleep(100 * time.Millisecond)
	return c.GetIncoming()
}
