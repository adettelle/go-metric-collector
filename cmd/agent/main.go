// Agent (HTTP-client) is used for collecting runtime metrics
// and their subsequent sending to the server by HTTP protocol.
package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	_ "net/http/pprof"

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

	mservice := metricservice.NewMetricService(config, metricAccumulator, client, 10)

	var wg sync.WaitGroup
	wg.Add(3)

	go mservice.SendLoop(time.Duration(config.ReportInterval), &wg)
	go mservice.RetrieveLoop(time.Duration(config.PollInterval), &wg)
	go mservice.AdditionalRetrieveLoop(time.Duration(config.PollInterval), &wg)

	if err = http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}
