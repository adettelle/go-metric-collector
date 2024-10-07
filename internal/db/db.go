package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/adettelle/go-metric-collector/pkg/retries"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func connect(dbParams string) (*sql.DB, error) {
	log.Println("Connecting to DB", dbParams)
	db, err := sql.Open("pgx", dbParams)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

type DBConnector interface {
	Connect() (*sql.DB, error)
}

type DBConnection struct {
	DBParams string
}

func NewDBConnection(params string) *DBConnection {
	return &DBConnection{
		DBParams: params,
	}
}

func (dbCon *DBConnection) Connect() (*sql.DB, error) {
	return retries.RunWithRetries("DB connection", 3,
		func() (*sql.DB, error) {
			return connect(dbCon.DBParams)
		},
		func(err error) bool {
			return true // все ошибки надо ретраить
		})
}
