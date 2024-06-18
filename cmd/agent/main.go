package main

import (
	"log"
	"sync"
	"time"

	"github.com/adettelle/go-metric-collector/internal/agent/config"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"

	"github.com/adettelle/go-metric-collector/internal/agent/metricservice"
)

func main() {
	metricsStorage := memstorage.New()
	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go metricservice.SendLoop(time.Duration(config.ReportInterval), metricsStorage, config.Address, &wg)
	go metricservice.RetrieveLoop(time.Duration(config.PollInterval), metricsStorage, &wg)
	wg.Wait()
}
