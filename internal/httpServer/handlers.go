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
		var body types.SubscriptionInfo
		err := decoder.Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		valErr := validate.Struct(body)
		if valErr != nil {
			http.Error(w, valErr.Error(), http.StatusBadRequest)
			return
		}
		options.SubscriptionChannel <- body
		w.WriteHeader(http.StatusOK)
	})
}

func ReadingsByTimeSpan(options HttpOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		deviceNameValid := false
		for _, val := range options.GlobalStore.Devices {
			if val.Device == vars["deviceName"] {
				deviceNameValid = true
				break
			}
		}
		if !deviceNameValid {
			http.Error(w, "no such device", http.StatusBadRequest)
			return
		}
	})
}
