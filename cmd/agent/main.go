// агент (HTTP-клиент) для сбора рантайм-метрик
// и их последующей отправки на сервер по протоколу HTTP
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

	mservice := metricservice.NewMetricService(config, metricsStorage)

	var wg sync.WaitGroup
	wg.Add(2)

	go mservice.SendLoop(time.Duration(config.ReportInterval), &wg)
	go mservice.RetrieveLoop(time.Duration(config.PollInterval), &wg)
	wg.Wait()
}
