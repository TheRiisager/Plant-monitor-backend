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
	"strconv"
	"sync"
	"syscall"
)

const SUB_FILE_PATH string = "./config/devices.json"

func main() {
	dbUrl := os.Getenv("DB_URL")
	maxWorkers, err := strconv.ParseInt(os.Getenv("MAX_DB_WORKERS"), 10, 64)
	if err != nil {
		panic(err)
	}

	context, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publisherAdded := make(chan types.DeviceInfo, 3)
	realtimeReadingsChannel := make(chan types.Reading, 3)

	globalStore := initGlobalStore(stop)

	database, err := database.MakeDatabaseWrapper(context, dbUrl, maxWorkers, globalStore)
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
				RealtimeChannel:     realtimeReadingsChannel,
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
				RealtimeChannel:     realtimeReadingsChannel,
				Database:            database,
				GlobalStore:         globalStore,
			})

	})
	wg.Wait()

	close(publisherAdded)
	close(realtimeReadingsChannel)

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
