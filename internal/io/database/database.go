package database

import (
	"context"
	"fmt"
	"riisager/backend_plant_monitor_go/internal/types"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DB_URL = ""

type DatabaseOptions struct {
	Context         context.Context
	DatabaseChannel chan types.Reading
}

func Run(options DatabaseOptions) {
	dbcontext, cancel := context.WithCancel(options.Context)

	dbpool, err := pgxpool.New(dbcontext, DB_URL)
	if err != nil {
		fmt.Println(err)
		cancel()
	}
	defer func() {
		dbpool.Close()
		cancel()
	}()

	for {
		select {
		case newReading := <-options.DatabaseChannel:
			go saveReading(newReading, dbpool, dbcontext)
		case <-dbcontext.Done():
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
		"INSERT INTO conditions VALUES (NOW(),"+
			reading.DeviceName+
			","+
			strconv.FormatFloat(float64(reading.Temperature), 'f', -1, 32)+
			"",
	)
	if err != nil {
		fmt.Println(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
