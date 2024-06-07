package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/adettelle/go-metric-collector/internal/handlers"
	store "github.com/adettelle/go-metric-collector/internal/storage"
	"github.com/go-chi/chi/v5"
)

type NetAddress struct {
	Host string
	Port int
}

func main() {
	addr := flag.String("a", "localhost:8080", "Net address localhost:port")
	flag.Parse()
	ensureAddrFLagIsCorrect(*addr)

	ms := store.NewMemStorage()
	mAPI := handlers.NewMetricAPI(ms)

	r := chi.NewRouter()

	// POST /update/counter/someMetric/123
	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", mAPI.CreateMetric)
	r.Get("/value/{metric_type}/{metric_name}", mAPI.GetMetricByValue)
	r.Get("/", mAPI.GetAllMetrics)

	fmt.Printf("Starting server on %s\n", *addr)

	err := http.ListenAndServe(*addr, r) // `:8080`
	if err != nil {
		log.Fatal(err)
	}
}

func ensureAddrFLagIsCorrect(addr string) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal(err)
	}

	_, err = strconv.Atoi(port)
	if err != nil {
		log.Fatal(fmt.Errorf("Invalid port: '%s'", port))
	}
}
