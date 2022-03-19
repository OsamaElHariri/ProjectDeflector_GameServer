package network

import "encoding/json"

type GameEndUserUpdate struct {
	PlayerId string `json:"playerId"`
	Games    int    `json:"games"`
	Wins     int    `json:"wins"`
}

func NotifyUserServiceGameEnd(updates []GameEndUserUpdate) {
	res, err := json.Marshal(map[string][]GameEndUserUpdate{
		"updates": updates,
	})

	if err != nil {
		return
	}

	SendPost("http://127.0.0.1:8080/users/internal/stats/games", res)
}
