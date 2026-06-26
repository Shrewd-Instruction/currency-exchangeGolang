package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var db *sql.DB

func connectDB(host, port, user, password, dbname string) error {
	connStr := fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("sqlserver", connStr)
	if err != nil {
		return fmt.Errorf("failed to open db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping db: %v", err)
	}

	return nil
}

func closeDB() {
	if db != nil {
		db.Close()
	}
}
