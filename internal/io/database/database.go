package database

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"riisager/backend_plant_monitor_go/internal/types"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/semaphore"
)

type DatabaseWrapper struct {
	Context     context.Context
	pool        *pgxpool.Pool
	globalStore *types.GlobalStore
	semaphore   *semaphore.Weighted
}

func MakeDatabaseWrapper(context context.Context, url string, maxWorkers int64, globalStore *types.GlobalStore) (DatabaseWrapper, error) {
	var database = DatabaseWrapper{}

	pool, err := pgxpool.New(context, url)
	if err != nil {
		fmt.Println(err)
		return database, err
	}
	database.Context = context
	database.pool = pool
	database.globalStore = globalStore
	database.semaphore = semaphore.NewWeighted(maxWorkers)

	return database, nil
}

func (db DatabaseWrapper) Close() {
	db.pool.Close()
}

func (db DatabaseWrapper) SaveReading(reading types.Reading) error {
	if err := db.semaphore.Acquire(db.Context, 1); err != nil {
		return err
	}
	defer db.semaphore.Release(1)

	_, err := db.pool.Exec(
		db.Context,
		"INSERT INTO readings VALUES (NOW(), $1, $2, $3)",
		reading.DeviceName,
		reading.Temperature,
		reading.SoilMoisture,
	)
	if err != nil {
		return err
	}

	return nil
}

func (db DatabaseWrapper) QueryTimeSpanByDevice(deviceName string, time string) ([]types.Reading, error) {
	db.globalStore.Mutex.RLock()
	//TODO make this a util function somewhere (where?)
	sliceIndex := slices.IndexFunc(db.globalStore.Devices, func(device types.DeviceInfo) bool {
		return device.Device == deviceName
	})
	if sliceIndex < 0 {
		return nil, errors.New("Device name is not valid!")
	}
	db.globalStore.Mutex.RUnlock()

	regex := regexp.MustCompile(`\b\d+\s*(?:second|minute|hour|day|week|month|year)s?\b`)
	time = regex.FindString(time)

	if time == "" {
		return nil, errors.New("invalid time string")
	}

	tx, err := db.pool.Begin(db.Context)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(db.Context)

	rows, err := tx.Query(
		db.Context,
		"SELECT * FROM readings WHERE device_name=$1 AND time > NOW() - INTERVAL $2 ORDER BY time DESC",
		deviceName,
		time,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(db.Context)
	if err != nil {
		return nil, err
	}

	readings, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Reading])
	if err != nil {
		return nil, err
	}

	return readings, nil
}
