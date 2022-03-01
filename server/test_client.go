package server

import (
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
	return <-c.Client.Incoming
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
	time.Sleep(time.Millisecond)
	return c.GetIncoming()
}

func (c *TestClient) DenyMatch(match int) common.Message {
	c.Client.Send(common.Message{
		Type: "match_declined",
		Payload: map[string]interface{}{
			"matchId": match,
		},
	})
	time.Sleep(time.Millisecond)
	return c.GetIncoming()
}
