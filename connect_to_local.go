package main

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func connectToLocal() (*sql.DB, error) {
	dbURI := "host=db port=5432 user=postgres password=password dbname=postgres sslmode=disable"

	dbPool, err := sql.Open("pgx", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	return dbPool, nil
}
