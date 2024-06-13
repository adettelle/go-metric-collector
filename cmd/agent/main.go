package main

import (
	"log"
	"sync"
	"time"

	"github.com/adettelle/go-metric-collector/internal/agent/config"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"

	"github.com/adettelle/go-metric-collector/internal/agent/metricservice"
)

// POLLINTERVAL=2 REPORTINTERVAL=10 go run ./cmd/agent/
func main() {
	metricsStorage := memstorage.New()
	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go metricservice.SendLoop(time.Duration(config.ReportInterval), metricsStorage, config.Address)
	go metricservice.RetrieveLoop(time.Duration(config.PollInterval), metricsStorage)
	wg.Wait()
}
