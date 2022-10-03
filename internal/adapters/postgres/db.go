package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Database struct {
	DB *pgxpool.Pool
}

func New(ctx context.Context, pgconn string) (*Database, error) {
	config, err := pgxpool.ParseConfig(pgconn)
	if err != nil {
		return nil, fmt.Errorf("postgres connection string parse failed: %v", err)
	}

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool failed: %v", err)
	}

	return &Database{DB: pool}, nil
}
