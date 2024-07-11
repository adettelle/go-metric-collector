// сервер для сбора рантайм-метрик, который собирает репорты от агентов по протоколу HTTP
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/adettelle/go-metric-collector/internal/api"
	"github.com/adettelle/go-metric-collector/internal/db"

	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/storage/dbstorage"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

func main() {

	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	var storager api.Storager
	var ms *memstorage.MemStorage

	if config.DBParams != "" {
		db, err := db.Connect(config.DBParams)
		if err != nil {
			log.Fatal(err)
		}
		storager = &dbstorage.DBStorage{
			Ctx: context.Background(),
			DB:  db,
		}
	} else {
		ms, err = memstorage.New(config.ShouldRestore(), config.StoragePath)
		if err != nil {
			log.Fatal(err)
		}

		storager = ms

		if config.StoreInterval > 0 {
			go memstorage.StartSaveLoop(time.Second*time.Duration(config.StoreInterval),
				config.StoragePath, ms)
		} else if config.StoreInterval == 0 {
			// если config.StoreInterval равен 0, то мы назначаем MemStorage FileName, чтобы
			// он мог синхронно писать изменения
			ms.FileName = config.StoragePath
		}
	}

	log.Println("config:", config)

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

		err = memstorage.WriteMetricsSnapshot(config.StoragePath, ms)
		if err != nil {
			log.Println("unable to write to file")
		}
		done <- true
	}()
	<-done
}
