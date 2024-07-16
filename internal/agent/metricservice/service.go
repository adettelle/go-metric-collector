// Сервисный слой отвечает за сбор и отправку метрик на удаленный сервер
package metricservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/adettelle/go-metric-collector/internal/agent/config"
	mstore "github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

// Структура MetricCollector получает и рассылает метрики, запускает свои циклы (Loop)
type MetricCollector struct {
	config *config.Config
	// store         StorageInterfase
	metricStorage *mstore.MemStorage
}

func NewMetricCollector(config *config.Config, metricStorage *mstore.MemStorage) *MetricCollector { // store StorageInterfase,

	return &MetricCollector{
		config: config,
		// store:         store,
		metricStorage: metricStorage,
	}
}

type MetricRequest struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (ms *MetricCollector) collectAllMetrics() ([]MetricRequest, error) {

	var metrics []MetricRequest

	for name, value := range ms.metricStorage.Gauge {
		metric := MetricRequest{
			MType: "gauge",
			ID:    name,
			Value: &value,
		}
		metrics = append(metrics, metric)
	}
	for name, delta := range ms.metricStorage.Counter {
		metric := MetricRequest{
			MType: "counter",
			ID:    name,
			Delta: &delta,
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// type MetricsRequest []MetricRequest

func (ms *MetricCollector) sendMultipleMetrics(metrics []MetricRequest) error {
	url := fmt.Sprintf("http://%s/updates/", ms.config.Address)

	chunks := rangeChunks(10, metrics)

	for i, chunk := range chunks {
		log.Printf("Sending chunk %d of %d, chunk size %d\n", i+1, len(chunks), len(chunk))

		data, err := json.Marshal(chunk)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
		if err != nil {
			return err
		}

		delay := 1 // попытки через 1, 3, 5 сек
		for i := 0; i < 4; i++ {
			log.Printf("Sending %d attempt", i)
			err = doSend(req)
			if err == nil {
				break
			} else {
				log.Printf("error while sending request: %v, is retriable: %v", err, isRetriableError(err))
				if i == 3 || !isRetriableError(err) {
					return err
				}
			}
			<-time.NewTicker(time.Duration(delay) * time.Second).C
			delay += 2
		}
		log.Printf("chunk %d sent successfully", i+1)
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

func doSend(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ue := UnsuccessfulStatusError{
			Message: fmt.Sprintf("response is not OK, status: %d", resp.StatusCode),
			Status:  resp.StatusCode, // статус, который пришел в ответе
		}
		// return fmt.Errorf("response is not OK, status: %d", resp.StatusCode)
		return &ue
	}
	return nil
}

func rangeChunks(chunkSize int, metrics []MetricRequest) [][]MetricRequest {

	res := [][]MetricRequest{}

	currentChunk := []MetricRequest{}

	for _, v := range metrics {
		currentChunk = append(currentChunk, v)
		if len(currentChunk) == chunkSize {
			res = append(res, currentChunk)
			currentChunk = []MetricRequest{}
		}
	}
	if len(currentChunk) > 0 {
		res = append(res, currentChunk)
	}
	return res
}

// sendLoop sends all metrics to the server (MemStorage) with delay
func (ms *MetricCollector) SendLoop(delay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for range ticker.C {
		log.Println("Sending metrics")
		// err := ms.sendAllMetrics()
		metrics, err := ms.collectAllMetrics() //
		if err != nil {
			log.Fatal(err)
		}
		err = ms.sendMultipleMetrics(metrics)
		if err != nil {
			log.Fatal(err) // паника после 3ей попытки или в случае не IsRetriableErr
		}
		ms.metricStorage.Reset()
	}
}

// retrieveLoop gets all metrics from MemStorage to the server with delay
func (ms *MetricCollector) RetrieveLoop(delay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for range ticker.C {
		log.Println("Retrieving metrics")
		ms.retrieveAllMetrics()
	}
}

// retrieveAllMetrics получает все метрики из пакета runtime
// и собирает дополнительные метрики (PollCount и RandomValue)
func (ms *MetricCollector) retrieveAllMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ms.metricStorage.AddCounterMetric("PollCount", 1)

	ms.metricStorage.AddGaugeMetric("RandomValue", rand.Float64())

	ms.metricStorage.AddGaugeMetric("Alloc", float64(m.Alloc))
	ms.metricStorage.AddGaugeMetric("BuckHashSys", float64(m.BuckHashSys))
	ms.metricStorage.AddGaugeMetric("Frees", float64(m.Frees))
	ms.metricStorage.AddGaugeMetric("GCCPUFraction", m.GCCPUFraction)
	ms.metricStorage.AddGaugeMetric("GCSys", float64(m.GCSys))
	ms.metricStorage.AddGaugeMetric("HeapAlloc", float64(m.HeapAlloc))
	ms.metricStorage.AddGaugeMetric("HeapIdle", float64(m.HeapIdle))
	ms.metricStorage.AddGaugeMetric("HeapInuse", float64(m.HeapInuse))
	ms.metricStorage.AddGaugeMetric("HeapObjects", float64(m.HeapObjects))
	ms.metricStorage.AddGaugeMetric("HeapReleased", float64(m.HeapReleased))
	ms.metricStorage.AddGaugeMetric("HeapSys", float64(m.HeapSys))
	ms.metricStorage.AddGaugeMetric("LastGC", float64(m.LastGC))
	ms.metricStorage.AddGaugeMetric("Lookups", float64(m.Lookups))
	ms.metricStorage.AddGaugeMetric("MCacheInuse", float64(m.MCacheInuse))
	ms.metricStorage.AddGaugeMetric("MCacheSys", float64(m.MCacheSys))
	ms.metricStorage.AddGaugeMetric("MSpanInuse", float64(m.MSpanInuse))
	ms.metricStorage.AddGaugeMetric("MSpanSys", float64(m.MSpanSys))
	ms.metricStorage.AddGaugeMetric("Mallocs", float64(m.Mallocs))
	ms.metricStorage.AddGaugeMetric("NextGC", float64(m.NextGC))
	ms.metricStorage.AddGaugeMetric("NumForcedGC", float64(m.NumForcedGC))
	ms.metricStorage.AddGaugeMetric("NumGC", float64(m.NumGC))
	ms.metricStorage.AddGaugeMetric("OtherSys", float64(m.OtherSys))
	ms.metricStorage.AddGaugeMetric("PauseTotalNs", float64(m.PauseTotalNs))
	ms.metricStorage.AddGaugeMetric("StackInuse", float64(m.StackInuse))
	ms.metricStorage.AddGaugeMetric("StackSys", float64(m.StackSys))
	ms.metricStorage.AddGaugeMetric("Sys", float64(m.Sys))
	ms.metricStorage.AddGaugeMetric("TotalAlloc", float64(m.TotalAlloc))
}
