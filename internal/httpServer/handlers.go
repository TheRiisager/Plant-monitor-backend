package httpServer

import (
	"encoding/json"
	"net/http"
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
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

	})
}
