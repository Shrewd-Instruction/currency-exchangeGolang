package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var DB *sql.DB

func ConnectDB(host, port, user, password, dbname string) error {
	connStr := fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("sqlserver", connStr)
	if err != nil {
		return fmt.Errorf("failed to open db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = DB.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping db: %v", err)
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
