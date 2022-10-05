package utils

import (
	"database/sql"
	"github.com/pressly/goose/v3"
)

func ApplyMigrations(connStr string, path string) error {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	err = goose.Up(conn, path)
	if err != nil {
		return err
	}
	return nil
}
