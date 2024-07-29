// агент (HTTP-клиент) для сбора рантайм-метрик
// и их последующей отправки на сервер по протоколу HTTP
package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adettelle/go-metric-collector/internal/agent/config"
	"github.com/adettelle/go-metric-collector/internal/agent/metrics"
	"github.com/adettelle/go-metric-collector/internal/agent/metricservice"
)

func main() {

	metricAccumulator := metrics.New()

	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Timeout: time.Second * 2, // интервал ожидания: 2 секунды
	}
	mservice := metricservice.NewMetricService(config, metricAccumulator, client)

	var wg sync.WaitGroup
	wg.Add(3)

	go mservice.SendLoop(time.Duration(config.ReportInterval), &wg)
	go mservice.RetrieveLoop(time.Duration(config.PollInterval), &wg)
	go mservice.AdditionalRetrieveLoop(time.Duration(config.PollInterval), &wg)
	wg.Wait()
}
