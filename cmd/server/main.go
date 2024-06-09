package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adettelle/go-metric-collector/internal/handlers"
	"github.com/adettelle/go-metric-collector/internal/server"
	store "github.com/adettelle/go-metric-collector/internal/storage"
	"github.com/go-chi/chi/v5"
)

type NetAddress struct {
	Host string
	Port int
}

func main() {

	config, err := server.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	// addr := flag.String("a", "localhost:8080", "Net address localhost:port")
	// flag.Parse()
	// ensureAddrFLagIsCorrect(*addr)

	ms := store.NewMemStorage()
	mAPI := handlers.NewMetricAPI(ms)

	r := chi.NewRouter()

	// POST /update/counter/someMetric/123
	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", mAPI.CreateMetric)
	r.Get("/value/{metric_type}/{metric_name}", mAPI.GetMetricByValue)
	r.Get("/", mAPI.GetAllMetrics)

	fmt.Printf("Starting server on %s\n", config.Address)

	err = http.ListenAndServe(config.Address, r) // `:8080`
	if err != nil {
		log.Fatal(err)
	}
}
