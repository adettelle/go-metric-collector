// Agent (HTTP-client) is used for collecting runtime metrics
// and their subsequent sending to the server by HTTP protocol.
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	caCert, err := os.ReadFile(config.ServerCert) // "./keys/server_cert.pem"
	if err != nil {
		log.Fatal("error in reading certificate: ", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// "./keys/client_cert.pem", "./keys/client_privatekey.pem"
	cert, err := tls.LoadX509KeyPair(config.ClientCert, config.CryptoKey)
	if err != nil {
		log.Fatal("error in loading key pair: ", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	mservice := metricservice.NewMetricService(config, metricAccumulator, client, 10)

	var wg sync.WaitGroup
	wg.Add(3)

	sendLoopTerm := make(chan struct{})
	go func() {
		if err = mservice.SendLoop(time.Duration(config.ReportInterval), &wg, sendLoopTerm); err != nil {
			log.Fatal(err)
		}
	}()
	retrievLoopTerm := make(chan struct{})
	go mservice.RetrieveLoop(time.Duration(config.PollInterval), &wg, retrievLoopTerm)
	additionalRetrieveLoopTerm := make(chan struct{})
	go mservice.AdditionalRetrieveLoop(time.Duration(config.PollInterval), &wg, additionalRetrieveLoopTerm)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		s := <-signals
		log.Printf("Got termination signal: %s. Graceful shutdown\n", s)
		retrievLoopTerm <- struct{}{}
		additionalRetrieveLoopTerm <- struct{}{}
		sendLoopTerm <- struct{}{}
	}()

	go func() {
		if err = http.ListenAndServe(":9000", nil); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}
