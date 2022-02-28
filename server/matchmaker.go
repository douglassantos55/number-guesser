package server

import "example.com/game/common"

type MatchMaker struct {
	currentId int
}

func NewMatchMaker() *MatchMaker {
	return &MatchMaker{
        currentId: 0,
    }
}

func (m *MatchMaker) Process(event Event, server *Server) {
	if event.Type == "match_found" {
		matchId := m.currentId + 1
		m.currentId = matchId

		event.Socket.WriteJSON(common.Message{
			Type: "confirm_match",
			Payload: map[string]interface{}{
				"matchId": matchId,
			},
		})
	}
}
