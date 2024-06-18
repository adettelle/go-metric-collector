// сервер для сбора рантайм-метрик, который собирает репорты от агентов по протоколу HTTP
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adettelle/go-metric-collector/internal/api"
	"github.com/adettelle/go-metric-collector/internal/handlers"
	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

func main() {

	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	ms := memstorage.New()
	mAPI := handlers.NewMetricHandlers(ms) // объект хэндлеров, ранее было handlers.NewMetricAPI(ms)
	r := api.NewMetricRouter(ms, mAPI)

	// r := chi.NewRouter()

	// // POST /update/counter/someMetric/123
	// r.Post("/update/{metric_type}/{metric_name}/{metric_value}", mAPI.CreateMetric)
	// r.Get("/value/{metric_type}/{metric_name}", mAPI.GetMetricByValue)
	// r.Get("/", mAPI.GetAllMetrics)

	fmt.Printf("Starting server on %s\n", config.Address)

	err = http.ListenAndServe(config.Address, r)
	if err != nil {
		log.Fatal(err)
	}
}
