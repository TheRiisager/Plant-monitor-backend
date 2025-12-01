package main

import (
	"context"
	"os"
	"os/signal"
	"riisager/backend_plant_monitor_go/internal/httpServer"
	"riisager/backend_plant_monitor_go/internal/mqtt"
	"riisager/backend_plant_monitor_go/internal/types"
	"syscall"
)

func main() {
	context, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publisherAdded := make(chan types.SubscriptionInfo, 1)

	go mqtt.Run(
		context,
		mqtt.MqttOptions{
			SubscriptionChannel: publisherAdded,
		})

	go httpServer.Run(
		httpServer.HttpOptions{
			Context:             context,
			SubscriptionChannel: publisherAdded,
		})

	<-context.Done()
}
