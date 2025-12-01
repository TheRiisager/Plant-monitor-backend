package httpServer

import (
	"context"
	"net/http"
	"riisager/backend_plant_monitor_go/internal/types"
)

type HttpOptions struct {
	Context             context.Context
	SubscriptionChannel chan types.SubscriptionInfo
}

func Run(options HttpOptions) {
	mux := http.NewServeMux()
	mux.Handle("/publisher/add", AddPublisher(options))

	http.ListenAndServe(":8080", mux)
}
