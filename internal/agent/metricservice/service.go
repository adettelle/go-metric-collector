// Сервисный слой отвечает за сбор и отправку метрик на удаленный сервер
package metricservice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/adettelle/go-metric-collector/internal/agent/config"
	"github.com/adettelle/go-metric-collector/pkg/collections"

	m "github.com/adettelle/go-metric-collector/internal/agent/metrics"
)

// MetricService structure receives and sends out metrics, runs its loops (Loop)
type MetricService struct {
	metricAccumulator *m.MetricAccumulator
	client            *Client
	rateLimit         int
	ChunkSize         int
}

func NewMetricService(
	config *config.Config,
	metricAccumulator *m.MetricAccumulator,
	client *http.Client,
	chunkSize int,
	// publicKey *rsa.PublicKey,
) *MetricService {
	return &MetricService{
		metricAccumulator: metricAccumulator,
		client: &Client{
			client:            client,
			url:               fmt.Sprintf("https://%s/updates/", config.Address),
			maxRequestRetries: config.MaxRequestRetries,
			encryptionKey:     config.Key,
			// publicKey:         publicKey,
		},
		rateLimit: config.RateLimit,
		ChunkSize: chunkSize,
	}
}

type MetricRequest struct {
	Delta *int64   `json:"delta,omitempty"` // metric's value when metric type is counter
	Value *float64 `json:"value,omitempty"` // metric's value when metric type is gauge
	ID    string   `json:"id"`              // metric's name
	MType string   `json:"type"`            // parameter that takes gauge or counter value
}

// SendLoop sends all metrics to the server (MemStorage) with delay
func (ms *MetricService) SendLoop(ctx context.Context, delay time.Duration, wg *sync.WaitGroup) error { // , term <-chan struct{}
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	chunks := make(chan []MetricRequest, ms.rateLimit) // 5
	results := make(chan bool)
	for i := 0; i < ms.rateLimit; i++ {
		go ms.StartWorker(i, chunks, results)
	}

	// в горутине вычитываем из канала results
	// (туда записывается успешная или неуспешная отправка чанка), освобождая его
	go func() {
		for res := range results {
			log.Printf("chunk sent result %t", res)
		}
	}()

	for {
		select {
		case <-ctx.Done(): // <-term:
			log.Println("Stopping SendLoop")
			return ms.finalizeSendLoop(chunks)
		case <-ticker.C:
			log.Println("Sending metrics")
			metrics, err := ms.collectAllMetrics()
			if err != nil {
				return err
			}
			err = ms.sendMultipleMetrics(metrics, chunks)
			if err != nil {
				return err // паника после 3ей попытки или в случае не IsRetriableErr
			}
			ms.metricAccumulator.Reset()
		}
	}
}

func (ms *MetricService) finalizeSendLoop(chunks chan []MetricRequest) error {
	log.Println("Sending final metrics")
	metrics, err := ms.collectAllMetrics()
	if err != nil {
		return err
	}
	err = ms.sendMultipleMetrics(metrics, chunks)
	if err != nil {
		return err // паника после 3ей попытки или в случае не IsRetriableErr
	}
	ms.metricAccumulator.Reset()
	return nil
}

// worker is our worker, which accepts two channels:
// jobs - task channel, it is the input data to be processed (входные данные для обработки)
// results - results channel, these are the results of the worker's work
func (ms *MetricService) StartWorker(id int, chunks <-chan []MetricRequest, results chan<- bool) {
	// worker:
	for chunk := range chunks {
		err := ms.client.SendMetricsChunk(id, chunk) // SendMetricsChunkEncrypted
		if err != nil {
			results <- false
		} else {
			results <- true
		}
	}
}

// RetrieveLoop collects all metrics from MemStorage.
// term - канал финилизации
func (ms *MetricService) RetrieveLoop(ctx context.Context, delay time.Duration, wg *sync.WaitGroup) { // , term <-chan struct{}
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping RetrieveLoop")
			return
		case <-ticker.C:
			log.Println("Retrieving metrics")
			RetrieveAllMetrics(ms.metricAccumulator)
		}
	}
}

// AdditionalRetrieveLoop gets aditional metrics from MemStorage to the server with delay.
func (ms *MetricService) AdditionalRetrieveLoop(ctx context.Context, delay time.Duration, wg *sync.WaitGroup) { // , term <-chan struct{}
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping AdditionalRetrieveLoop")
			return
		case <-ticker.C:
			log.Println("Retrieving additional metrics")
			retrieveAdditionalGaugeMetrics(ms.metricAccumulator)
		}
	}
}

func (ms *MetricService) collectAllMetrics() ([]MetricRequest, error) {
	var metrics []MetricRequest

	gaugeMetrics := ms.metricAccumulator.GetAllGaugeMetrics()

	for name, value := range gaugeMetrics {
		metric := MetricRequest{
			MType: "gauge",
			ID:    name,
			Value: &value,
		}
		metrics = append(metrics, metric)
	}

	counterMetrics := ms.metricAccumulator.GetAllCounterMetrics()

	for name, delta := range counterMetrics {
		metric := MetricRequest{
			MType: "counter",
			ID:    name,
			Delta: &delta,
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (ms *MetricService) sendMultipleMetrics(metrics []MetricRequest,
	workerRequests chan<- []MetricRequest) error {
	// url := fmt.Sprintf("http://%s/updates/", ms.config.Address)

	chunks := collections.RangeChunks(ms.ChunkSize, metrics) // 10

	for _, chunk := range chunks {
		workerRequests <- chunk
	}
	return nil
}

type UnsuccessfulStatusError struct {
	Message string
	Status  int
}

func (ue UnsuccessfulStatusError) Error() string {
	return ue.Message
}

// будем считать, что стоит повторить запрос, если у нас произошла проблема с запросом (Client.Do)
// это мб. проблема с сетью, либо если у нас пришел ответ со статусом 500,
// то есть сервер возможно сможет обработать в следующий раз
func isRetriableError(err error) bool {
	var statusErr *UnsuccessfulStatusError
	if errors.As(err, &statusErr) {
		return statusErr.Status == http.StatusInternalServerError
	}
	return true
}
