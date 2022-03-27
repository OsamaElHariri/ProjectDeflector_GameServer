package network

import (
	"encoding/json"
	"os"
)

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

	SendPost(os.Getenv("INTERNAL_SERVICES_URL")+"/users/internal/stats/games", res)
}
