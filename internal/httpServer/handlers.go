package httpServer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/SierraSoftworks/multicast/v2"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func AddPublisher(options HttpOptions) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validate := validator.New(validator.WithRequiredStructEnabled())
		decoder := json.NewDecoder(r.Body)
		var body types.DeviceInfo
		err := decoder.Decode(&body)
		if err != nil {
			http.Error(w, "request format incorrect", http.StatusBadRequest)
			return
		}

		err = validate.Struct(body)
		if err != nil {
			http.Error(w, "request format incorrect", http.StatusBadRequest)
			return
		}
		options.SubscriptionChannel <- body
		w.WriteHeader(http.StatusOK)
	})
}

func ReadingsByTimeSpan(options HttpOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validate := validator.New(validator.WithRequiredStructEnabled())
		decoder := json.NewDecoder(r.Body)
		var body QueryTimeSpanByDeviceRequest
		err := decoder.Decode(&body)
		if err != nil {
			http.Error(w, "request format incorrect", http.StatusBadRequest)
			return
		}

		err = validate.Struct(body)
		if err != nil {
			http.Error(w, "request format incorrect", http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		result, err := options.Database.QueryTimeSpanByDevice(vars["deviceName"], body.Timespan)
		if err != nil {
			http.Error(w, "an error occured", http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(result)
		if err != nil {
			http.Error(w, "an error occured", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})
}

func websocketRealTimeReadings(options HttpOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deviceName := mux.Vars(r)["deviceName"]

		if !options.GlobalStore.DeviceExists(deviceName) {
			http.Error(w, "device does not exist", http.StatusBadRequest)
			return
		}

		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "failed to upgrade connection", http.StatusBadRequest)
			return
		}
		defer conn.Close()

		listener := multicast.NewListener(options.RealtimeChannel)
		go websocketWriter(conn, listener, deviceName)
		websocketReader(conn)
	})
}
