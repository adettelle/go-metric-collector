// сервер для сбора рантайм-метрик, который собирает репорты от агентов по протоколу HTTP
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/adettelle/go-metric-collector/internal/api"
	database "github.com/adettelle/go-metric-collector/internal/db"

	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/storage/dbstorage"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

func main() {

	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("config:", config)

	storager, err := initStorager(config)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		fmt.Printf("Starting server on %s\n", config.Address)
		mAPI := api.NewMetricHandlers(storager, config) // объект хэндлеров, ранее было handlers.NewMetricAPI(ms)
		router := api.NewMetricRouter(storager, mAPI)
		err := http.ListenAndServe(config.Address, router)
		if err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	done := make(chan bool, 1)

	go func() {

		s := <-c
		log.Printf("Got termination signal: %s. Graceful shutdown", s)

		err = storager.Finalize() // memstorage.WriteMetricsSnapshot(config.StoragePath, ms)
		if err != nil {
			log.Println("unable to write to file")
		}
		done <- true
	}()
	<-done
}

// init потому что он не только конструирует, но и запускает сопутсвующие процессы
// в зависимости от того, какой storager мы выбрали
func initStorager(config *config.Config) (api.Storager, error) {
	var storager api.Storager

	if config.DBParams != "" {
		db, err := database.ConnectWithRerties(config.DBParams)
		if err != nil {
			log.Fatal(err)
		}
		storager = &dbstorage.DBStorage{
			Ctx: context.Background(),
			DB:  db,
		}

		err = database.CreateTable(db, context.Background())
		if err != nil {
			log.Fatal(err)
		}
	} else {
		var ms *memstorage.MemStorage
		ms, err := memstorage.New(config.ShouldRestore(), config.StoragePath, config.StoreInterval)
		if err != nil {
			return nil, err
		}

		storager = ms

	}
	return storager, nil
}
