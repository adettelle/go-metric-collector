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

func connect(dbParams string) (*sql.DB, error) {
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

func ConnectWithRerties(dbParams string) (*sql.DB, error) {
	delay := 1 // попытки через 1, 3, 5 сек
	for i := 0; i < 4; i++ {
		log.Printf("Sending %d attempt", i)
		db, err := connect(dbParams)
		if err == nil {
			return db, nil
		} else {
			log.Printf("error while connecting to db: %v", err)
			if i == 3 {
				return nil, err
			}
		}
		<-time.NewTicker(time.Duration(delay) * time.Second).C
		delay += 2
	}

	return nil, nil
}
