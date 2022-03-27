package network

import (
	"encoding/json"
	"os"
)

func SocketBroadcast(ids []string, event string, payload map[string]interface{}) error {
	res, err := json.Marshal(map[string]interface{}{
		"event":   event,
		"payload": payload,
	})

	if err != nil {
		return err
	}

	for _, id := range ids {
		go socketBroadcast(id, res)
	}
	return nil
}

func socketBroadcast(id string, payload []byte) {
	SendPost(os.Getenv("INTERNAL_SERVICES_URL")+"/realtime/internal/notify/"+id, payload)
}
