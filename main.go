package main

import (
	"context"
	"os"
	"os/signal"
	"riisager/backend_plant_monitor_go/internal/httpServer"
	"riisager/backend_plant_monitor_go/internal/mqtt"
	"riisager/backend_plant_monitor_go/internal/types"
	"sync"
	"syscall"
)

func main() {
	context, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publisherAdded := make(chan types.SubscriptionInfo, 1)

	var wg sync.WaitGroup
	wg.Go(func() {
		defer wg.Done()
		defer stop()
		mqtt.Run(
			mqtt.MqttOptions{
				Context:             context,
				SubscriptionChannel: publisherAdded,
			})

	})
	wg.Go(func() {
		defer wg.Done()
		defer stop()
		httpServer.Run(
			httpServer.HttpOptions{
				Context:             context,
				SubscriptionChannel: publisherAdded,
			})

	})

	wg.Wait()
}
