package database

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/semaphore"
)

type DatabaseWrapper struct {
	Context   context.Context
	pool      *pgxpool.Pool
	semaphore *semaphore.Weighted
}

func MakeDatabaseWrapper(context context.Context, url string, maxWorkers int64) (DatabaseWrapper, error) {
	var database = DatabaseWrapper{}

	pool, err := pgxpool.New(context, url)
	if err != nil {
		fmt.Println(err)
		return database, err
	}
	database.Context = context
	database.pool = pool
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
	go func() {
		tx, err := db.pool.Begin(db.Context)
		if err != nil {
			return
		}
		defer tx.Rollback(db.Context)

		_, err = tx.Exec(
			db.Context,
			"INSERT INTO readings VALUES (NOW(), $1, $2, $3)",
			reading.DeviceName,
			reading.Temperature,
			reading.SoilMoisture,
		)
		if err != nil {
			return
		}

		err = tx.Commit(db.Context)
		if err != nil {
			return
		}
	}()

	return nil
}

func (db DatabaseWrapper) QueryTimeSpanByDevice(deviceName string, time string) ([]types.Reading, error) {
	//TODO validate device name

	regex := regexp.MustCompile(`\b\d+\s*(?:second|minute|hour|day|week|month|year)s?\b`)

	if !regex.MatchString(time) {
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
