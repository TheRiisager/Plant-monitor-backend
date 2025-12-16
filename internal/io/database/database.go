package database

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"riisager/backend_plant_monitor_go/internal/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseOptions struct {
	Context         context.Context
	DatabaseChannel chan types.Reading
	Dbpool          *pgxpool.Pool
}

func Run(options DatabaseOptions) {

	for {
		select {
		case newReading := <-options.DatabaseChannel:
			go saveReading(newReading, options.Dbpool, options.Context)
		case <-options.Context.Done():
			return
		}
	}
}

func saveReading(reading types.Reading, dbpool *pgxpool.Pool, ctx context.Context) {
	tx, err := dbpool.Begin(ctx)
	if err != nil {
		fmt.Println(err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		"INSERT INTO readings VALUES (NOW(), $1, $2, $3)",
		reading.DeviceName,
		reading.Temperature,
		reading.SoilMoisture,
	)
	if err != nil {
		fmt.Println(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func QueryTimeSpanByDevice(dbpool *pgxpool.Pool, deviceName string, time string, ctx context.Context) ([]types.Reading, error) {
	//TODO validate device name

	regex := regexp.MustCompile(`\b\d+\s*(?:second|minute|hour|day|week|month|year)s?\b`)

	if !regex.MatchString(time) {
		return nil, errors.New("invalid time string")
	}

	tx, err := dbpool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(
		ctx,
		"SELECT * FROM readings WHERE device_name=$1 AND time > NOW() - INTERVAL $2 ORDER BY time DESC",
		deviceName,
		time,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	readings, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Reading])
	if err != nil {
		return nil, err
	}

	return readings, nil
}
