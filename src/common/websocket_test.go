package common

import (
	"log"
	"testing"
	"time"
)

func Test_websocket(t *testing.T) {
	URL := "ws://127.0.0.1:8888/api/kernels/059bb989-7696-4dec-870f-033ff2a2c968/channels"

	wsClient := NewWebSocketClient(URL)
	defer wsClient.Close()
	err := wsClient.Activate()
	if err != nil {
		log.Printf("Cannot connect to the websocket: %v", err)
		return
	}
	idleCount := 0
	for {
		select {
		case message := <-wsClient.ResultChan:

			if InfoRequestResult(message, &idleCount) {
				log.Println("active the kernel done")
				return
			}

		case <-time.After(3 * time.Second):
			log.Printf("Waiting Timeout")
			return
		}
	}

}
