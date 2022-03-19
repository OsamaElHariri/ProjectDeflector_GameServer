package network

import (
	"bytes"
	"net/http"
)

func SendPost(url string, payload []byte) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err == nil {
		resp.Body.Close()
	}
}
