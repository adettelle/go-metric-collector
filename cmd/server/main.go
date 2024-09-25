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
	database "github.com/adettelle/go-metric-collector/internal/db"
	"github.com/adettelle/go-metric-collector/internal/migrator"

	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/storage/dbstorage"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("config:", cfg)

	storager, err := initStorager(cfg)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		fmt.Printf("Starting server on %s\n", cfg.Address)
		mAPI := api.NewMetricHandlers(storager, cfg) // объект хэндлеров, ранее было handlers.NewMetricAPI(ms)
		router := api.NewMetricRouter(storager, mAPI)
		err := http.ListenAndServe(cfg.Address, router)
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
			log.Println(err)
			log.Println("unable to write to file")
		}
		done <- true
	}()
	<-done
}

// init потому что он не только конструирует, но и запускает сопутсвующие процессы
// в зависимости от того, какой storager мы выбрали
func initStorager(cfg *config.Config) (api.Storager, error) {
	log.Println("config in initStorager:", cfg)
	var storager api.Storager

	if cfg.DBParams != "" {

		db, err := database.NewDBConnection(cfg.DBParams).Connect()
		if err != nil {
			log.Fatal(err)
		}
		storager = &dbstorage.DBStorage{
			Ctx: context.Background(),
			DB:  db,
		}

		migrator.MustApplyMigrations(cfg.DBParams)
		// err = database.CreateTable(db, context.Background())
		// if err != nil {
		// 	log.Fatal(err)
		// }
	} else {
		var ms *memstorage.MemStorage
		log.Println("cfg.StoragePath in initStorager:", cfg.StoragePath)
		ms, err := memstorage.New(cfg.ShouldRestore(), cfg.StoragePath)
		log.Println("ms.StoragePath in initStorager:", ms.FileName)
		if err != nil {
			return nil, err
		}

		if cfg.StoreInterval > 0 {
			go memstorage.StartSaveLoop(time.Second*time.Duration(cfg.StoreInterval),
				cfg.StoragePath, ms)
		} else if cfg.StoreInterval == 0 {
			// если config.StoreInterval равен 0, то мы назначаем MemStorage FileName, чтобы
			// он мог синхронно писать изменения
			ms.FileName = cfg.StoragePath
		}

		storager = ms

	}
	return storager, nil
}
