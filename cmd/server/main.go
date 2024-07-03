// сервер для сбора рантайм-метрик, который собирает репорты от агентов по протоколу HTTP
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/adettelle/go-metric-collector/internal/api"
	"github.com/adettelle/go-metric-collector/internal/handlers"

	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

func main() {

	// var ms *memstorage.MemStorage
	// var err error

	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	ms, err := memstorage.New(config.ShouldRestore(), config.StoragePath)
	if err != nil {
		log.Fatal(err)
	}

	if config.StoreInterval > 0 {
		go memstorage.StartSaveLoop(time.Second*time.Duration(config.StoreInterval),
			config.StoragePath, ms)
	} else if config.StoreInterval == 0 {
		// если config.StoreInterval равен 0, то мы назначаем MemStorage FileName, чтобы
		// он мог синхронно писать изменения
		ms.FileName = config.StoragePath
	}

	log.Println("config:", config)
	go startServer(config, ms)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	log.Printf("Got termination signal: %s. Graceful shutdown", s)

	err = memstorage.WriteMetricsSnapshot(config.StoragePath, ms)
	if err != nil {
		log.Println("unable to write to file")
	}
}

func startServer(config *config.Config, ms *memstorage.MemStorage) {
	fmt.Printf("Starting server on %s\n", config.Address)
	mAPI := handlers.NewMetricHandlers(ms, config) // объект хэндлеров, ранее было handlers.NewMetricAPI(ms)
	r := api.NewMetricRouter(ms, mAPI)
	err := http.ListenAndServe(config.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
