package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func CreateTable(db *sql.DB, ctx context.Context) error { // metric_type_enum

	sqlStqtement := "create table if not exists metric" +
		"(id serial primary key , metric_type text not null," +
		"metric_id varchar(30) not null, value double precision not null default 0," +
		"delta integer not null default 0, created_at timestamp not null default now());"

	_, err := db.ExecContext(ctx, sqlStqtement)
	if err != nil {
		return err
	}

	return nil
}

func Connect(dbParams string) (*sql.DB, error) {
	log.Println("Connecting to DB", dbParams)
	db, err := sql.Open("pgx", dbParams)
	if err != nil {
		return nil, err
	}
	// defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
