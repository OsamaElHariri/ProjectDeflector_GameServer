package broadcast

import (
	"bytes"
	"encoding/json"
	"net/http"
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
	resp, _ := http.Post("http://127.0.0.1:8080/realtime/notify/"+id, "application/json", bytes.NewBuffer(payload))
	resp.Body.Close()
}
