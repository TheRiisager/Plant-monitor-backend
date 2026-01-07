package httpServer

import (
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/SierraSoftworks/multicast/v2"
	"github.com/gorilla/websocket"
)

func websocketReader(conn *websocket.Conn) {
	for {
		if _, _, err := conn.NextReader(); err != nil {
			conn.Close()
			break
		}
	}
}

func websocketWriter(conn *websocket.Conn, channel *multicast.Listener[types.Reading], deviceName string) {
	defer conn.Close()

	for reading := range channel.C {
		if reading.DeviceName == deviceName {
			conn.WriteJSON(reading)
		}
	}
}
