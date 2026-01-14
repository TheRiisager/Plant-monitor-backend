package httpServer

import (
	"context"
	"net/http"
	"riisager/backend_plant_monitor_go/internal/io/database"
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/gorilla/mux"
)

type HttpOptions struct {
	Context             context.Context
	SubscriptionChannel chan<- types.DeviceInfo
	RealtimeChannel     <-chan types.Reading
	Database            database.DatabaseWrapper
	GlobalStore         *types.GlobalStore
}

func Run(options HttpOptions) {
	mux := mux.NewRouter()
	mux.Handle("/publisher", AddPublisher(options)).Methods("POST")
	mux.Handle("/readings/{deviceName}", ReadingsByTimeSpan(options)).Methods("GET")
	mux.Handle("/readings/{deviceName}/realtime", websocketRealTimeReadings(options))

	http.ListenAndServe(":8080", mux)
}
