package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectDb(ctx context.Context, dbUrl string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)

	db.SetConnMaxIdleTime(time.Minute * 5)
	db.SetConnMaxLifetime(time.Minute * 30)

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}