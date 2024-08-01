// Сервисный слой отвечает за сбор и отправку метрик на удаленный сервер
package metricservice

import (
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

// Структура MetricService получает и рассылает метрики, запускает свои циклы (Loop)
type MetricService struct {
	metricAccumulator *m.MetricAccumulator
	rateLimit         int
	client            *Client
}

func NewMetricService(config *config.Config, metricAccumulator *m.MetricAccumulator, client *http.Client) *MetricService { // store StorageInterfase,

	return &MetricService{
		metricAccumulator: metricAccumulator,
		client: &Client{
			client:            client,
			url:               fmt.Sprintf("http://%s/updates/", config.Address),
			maxRequestRetries: config.MaxRequestRetries,
			encryptionKey:     config.Key,
		},
		rateLimit: config.RateLimit,
	}
}

type MetricRequest struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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

	chunks := collections.RangeChunks(10, metrics)

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

// sendLoop sends all metrics to the server (MemStorage) with delay
func (ms *MetricService) SendLoop(delay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	chunks := make(chan []MetricRequest, ms.rateLimit) // 5
	results := make(chan bool)
	for i := 0; i < ms.rateLimit; i++ {
		go ms.StartWorker(i, chunks, results)
	}

	go func() {
		for res := range results {
			log.Printf("FIXME: chunk sent result %t", res)
		}
	}()

	for range ticker.C {
		log.Println("Sending metrics")
		metrics, err := ms.collectAllMetrics()
		if err != nil {
			log.Fatal(err)
		}
		err = ms.sendMultipleMetrics(metrics, chunks)
		if err != nil {
			log.Fatal(err) // паника после 3ей попытки или в случае не IsRetriableErr
		}
		ms.metricAccumulator.Reset()
	}
}

// worker это наш рабочий, который принимает два канала:
// jobs - канал задач, это входные данные для обработки
// results - канал результатов, это результаты работы воркера
func (ms *MetricService) StartWorker(id int, chunks <-chan []MetricRequest, results chan<- bool) {
	// worker:
	for chunk := range chunks {
		err := ms.client.SendMetricsChunk(id, chunk)
		if err != nil {
			results <- false
		} else {
			results <- true
		}
	}
}

// retrieveLoop gets all metrics from MemStorage to the server with delay
func (ms *MetricService) RetrieveLoop(delay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for range ticker.C {
		log.Println("Retrieving metrics")
		retrieveAllMetrics(ms.metricAccumulator)
	}
}

// AdditionalRetrieveLoop gets aditional metrics from MemStorage to the server with delay
func (ms *MetricService) AdditionalRetrieveLoop(delay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for range ticker.C {
		log.Println("Retrieving additional metrics")
		retrieveAdditionalGaugeMetrics(ms.metricAccumulator)
	}
}
