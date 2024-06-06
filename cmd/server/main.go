package main

import (
	"log"
	"net/http"

	"github.com/adettelle/go-metric-collector/internal/handlers"
	store "github.com/adettelle/go-metric-collector/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	ms := store.NewMemStorage()
	mAPI := handlers.NewMetricAPI(ms)

	r := chi.NewRouter()

	// POST /update/counter/someMetric/123
	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", mAPI.CreateMetric)
	r.Get("/value/{metric_type}/{metric_name}", mAPI.GetMetricByValue)
	r.Get("/", mAPI.GetAllMetrics)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		log.Fatal(err)
	}
}
