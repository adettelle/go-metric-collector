// сервер для сбора рантайм-метрик, который собирает репорты от агентов по протоколу HTTP
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adettelle/go-metric-collector/internal/api"
	"github.com/adettelle/go-metric-collector/internal/handlers"

	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

func main() {

	var ms *memstorage.MemStorage
	var err error

	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	fi, err := os.Stat("/tmp/metrics-db.json")
	if err != nil {
		fmt.Println(err)
	}

	fileStoragePath, err := os.Stat(config.StoragePath)
	if err != nil {
		log.Printf("No such file: %v", config.StoragePath) // ?????????
	}

	if _, err := os.Stat("/tmp/metrics-db.json"); os.IsNotExist(err) || fi.Size() == 0 {
		if config.StoragePath == "/tmp/metrics-db.json" && config.Restore {
			config.Restore = false
		}
	}

	if config.Restore {
		if fileStoragePath.Size() == 0 {
			ms = &memstorage.MemStorage{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{},
			}
		} else {
			ms, err = memstorage.ReadMetricsSnapshot(config.StoragePath)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		ms = memstorage.New()
	}
	fmt.Println("ms:", ms.Counter, ms.Gauge)

	if config.StoreInterval > 0 {
		go memstorage.StartSaveLoop(time.Second*time.Duration(config.StoreInterval),
			config.StoragePath, ms)
	} else if config.StoreInterval == 0 {
		// если config.StoreInterval равен 0, то мы назначаем MemStorage FileName, чтобы
		// он мог синхронно писать изменения
		ms.FileName = config.StoragePath
	}

	mAPI := handlers.NewMetricHandlers(ms, config) // объект хэндлеров, ранее было handlers.NewMetricAPI(ms)
	r := api.NewMetricRouter(ms, mAPI)

	fmt.Printf("Starting server on %s\n", config.Address)

	err = http.ListenAndServe(config.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
