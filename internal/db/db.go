package db

import (
	"context"
	"github.com/jackc/pgx/v5"
)

func CheckConnectivity(pgURL string) error {
	conn, err := pgx.Connect(context.Background(), pgURL)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	return conn.Ping(context.Background())
}
