package main

import (
	"log"
	"net/http"

	"github.com/adettelle/go-metric-collector/internal/handlers"
	store "github.com/adettelle/go-metric-collector/internal/storage"
)

var storage = store.NewMemStorage()

func main() {
	ms := store.NewMemStorage()
	mAPI := handlers.NewMetricApi(ms)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{metric_type}/{metric_name}/{metric_value}", mAPI.CreateMetric)
	//http.HandleFunc("POST /update/{metric_type}/{metric_name}/{metric_value}", CreateMetric)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		log.Fatal(err)
	}
}
