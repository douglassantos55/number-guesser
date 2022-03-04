package server

import (
	"time"

	"example.com/game/client/client"
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

func (c *TestClient) GetIncoming() client.Message {
	return <-c.Client.Incoming
}

func (c *TestClient) QueueUp() client.Message {
	c.Client.Send(client.Message{
		Type: "queue_up",
	})
	return c.GetIncoming()
}

func (c *TestClient) Guess(guess, gameId int) client.Message {
	c.Client.Send(client.Message{
		Type: "guess",
		Payload: map[string]interface{}{
			"guess":  guess,
			"gameId": gameId,
		},
	})
	return c.GetIncoming()
}

func (c *TestClient) AcceptMatch(match int) client.Message {
	c.Client.Send(client.Message{
		Type: "match_confirmed",
		Payload: map[string]interface{}{
			"matchId": match,
		},
	})
	time.Sleep(time.Millisecond)
	return c.GetIncoming()
}

func (c *TestClient) DenyMatch(match int) client.Message {
	c.Client.Send(client.Message{
		Type: "match_declined",
		Payload: map[string]interface{}{
			"matchId": match,
		},
	})
	time.Sleep(time.Millisecond)
	return c.GetIncoming()
}
