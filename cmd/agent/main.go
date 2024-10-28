// Agent (HTTP-client) is used for collecting runtime metrics
// and their subsequent sending to the server by HTTP protocol.
package main

import (
	"context"
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

	grpcUrl := config.GrpcUrl
	var mservice *metricservice.MetricService
	var sender metricservice.MetricSender

	if grpcUrl != "" {
		sender = metricservice.NewGrpcSender(grpcUrl)
	} else {
		sender = metricservice.NewHTTPSender(client, fmt.Sprintf("https://%s/updates/", config.Address), config.MaxRequestRetries, config.Key)
	}

	mservice = metricservice.NewMetricService(config, metricAccumulator, sender, 10)

	var wg sync.WaitGroup
	wg.Add(3)

	sendLoopCtxWithCancel, cancelSendLoop := context.WithCancel(context.Background())

	go func() {
		if err = mservice.SendLoop(sendLoopCtxWithCancel, time.Duration(config.ReportInterval), &wg); err != nil {
			log.Fatal(err)
		}
	}()

	retrieveLoopCtxWithCancel, cancelRetrieveLoop := context.WithCancel(context.Background())
	go mservice.RetrieveLoop(retrieveLoopCtxWithCancel, time.Duration(config.PollInterval), &wg)

	addRetrieveLoopCtxWithCancel, cancelAddRetrieveLoop := context.WithCancel(context.Background())
	go mservice.AdditionalRetrieveLoop(addRetrieveLoopCtxWithCancel, time.Duration(config.PollInterval), &wg)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		s := <-signals
		log.Printf("Got termination signal: %s. Graceful shutdown\n", s)

		cancelRetrieveLoop()
		cancelAddRetrieveLoop()
		cancelSendLoop()
	}()

	startProfiling()

	wg.Wait()
}

func startProfiling() {
	go func() {
		if err := http.ListenAndServe(":9000", nil); err != nil {
			log.Fatal(err)
		}
	}()
}
