package common

type Processor interface {
	execute(msg Message) Response
}

type Message struct {
	Type    string
	Payload map[string]interface{}
}

func QueueUp() Message {
	return Message{
		Type: "queue_up",
	}
}
