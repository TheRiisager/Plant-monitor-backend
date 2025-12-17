package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"riisager/backend_plant_monitor_go/internal/httpServer"
	"riisager/backend_plant_monitor_go/internal/io/database"
	"riisager/backend_plant_monitor_go/internal/io/json"
	"riisager/backend_plant_monitor_go/internal/mqtt"
	"riisager/backend_plant_monitor_go/internal/types"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DB_URL = ""
const SUB_FILE_PATH string = "./config/subscriptions.json"

func main() {

	//todo maybe add logic to clean database if devices have been manually removed from the json? Not sure if necessary

	context, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publisherAdded := make(chan types.SubscriptionInfo, 1)

	database, err := database.MakeDatabaseWrapper(context, DB_URL)
	if err != nil {
		fmt.Println(err)
		stop()
	}
	defer database.Close()

	globalStore := initGlobalStore(stop)

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
}

func initDatabase(ctx context.Context, stop context.CancelFunc) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, DB_URL)
	if err != nil {
		fmt.Println(err)
		stop()
		return nil
	}
	return pool
}

func initGlobalStore(stop context.CancelFunc) *types.GlobalStore {
	devices, err := json.ReadfromFile[[]types.SubscriptionInfo](SUB_FILE_PATH)
	if err != nil {
		stop()
		return nil
	}
	return &types.GlobalStore{
		Devices: devices,
	}
}
