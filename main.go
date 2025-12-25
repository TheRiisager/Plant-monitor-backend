package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"riisager/backend_plant_monitor_go/internal/httpServer"
	"riisager/backend_plant_monitor_go/internal/io/database"
	"riisager/backend_plant_monitor_go/internal/io/file"
	"riisager/backend_plant_monitor_go/internal/mqtt"
	"riisager/backend_plant_monitor_go/internal/types"
	"sync"
	"syscall"
)

const DB_URL = ""
const SUB_FILE_PATH string = "./config/subscriptions.json"
const MAX_WORKERS int64 = 10

func main() {

	//todo maybe add logic to clean database if devices have been manually removed from the json? Not sure if necessary

	context, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publisherAdded := make(chan types.DeviceInfo, 1)

	globalStore := initGlobalStore(stop)

	database, err := database.MakeDatabaseWrapper(context, DB_URL, MAX_WORKERS, globalStore)
	if err != nil {
		fmt.Println(err)
		stop()
	}
	defer database.Close()

	var wg sync.WaitGroup
	wg.Go(func() {
		defer wg.Done()
		defer stop()
		mqtt.Run(
			mqtt.MqttOptions{
				Context:             context,
				SubscriptionChannel: publisherAdded,
				Database:            database,
				GlobalStore:         globalStore,
			})

	})
	wg.Go(func() {
		defer wg.Done()
		defer stop()
		httpServer.Run(
			httpServer.HttpOptions{
				Context:             context,
				SubscriptionChannel: publisherAdded,
				Database:            database,
				GlobalStore:         globalStore,
			})

	})
	wg.Wait()

	globalStore.Mutex.RLock()
	err = file.WriteToFile(
		SUB_FILE_PATH,
		&globalStore.Devices,
	)

	if err != nil {
		fmt.Println(err)
	}
	globalStore.Mutex.RUnlock()
}

func initGlobalStore(stop context.CancelFunc) *types.GlobalStore {
	devices, err := file.ReadfromFile[[]types.DeviceInfo](SUB_FILE_PATH)
	if err != nil {
		stop()
		return nil
	}
	return &types.GlobalStore{
		Devices: devices,
	}
}
