package httpServer

import (
	"context"
	"net/http"
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HttpOptions struct {
	Context             context.Context
	SubscriptionChannel chan types.SubscriptionInfo
	Dbpool              *pgxpool.Pool
	GlobalStore         *types.GlobalStore
}

func Run(options HttpOptions) {
	mux := mux.NewRouter()
	mux.Handle("/publisher", AddPublisher(options)).Methods("POST")
	mux.Handle("/readings/{deviceName}", ReadingsByTimeSpan(options)).Methods("GET")

	http.ListenAndServe(":8080", mux)
}
