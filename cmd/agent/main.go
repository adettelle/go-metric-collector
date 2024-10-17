// Agent (HTTP-client) is used for collecting runtime metrics
// and their subsequent sending to the server by HTTP protocol.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/adettelle/go-metric-collector/internal/agent/config"
	"github.com/adettelle/go-metric-collector/internal/agent/metrics"
	"github.com/adettelle/go-metric-collector/internal/agent/metricservice"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	var err error

	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)

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

	go func() {
		if err = mservice.SendLoop(time.Duration(config.ReportInterval), &wg); err != nil {
			log.Fatal(err)
		}
	}()
	go mservice.RetrieveLoop(time.Duration(config.PollInterval), &wg)
	go mservice.AdditionalRetrieveLoop(time.Duration(config.PollInterval), &wg)

	if err = http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}

	wg.Wait()
}
