// Сервисный слой отвечает за сбор и отправку метрик на удаленный сервер
package metricservice

import (
	"bytes"
	"encoding/json"
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

type MetricsRequest struct {
	Metrics []MetricRequest
}

func (ms *MetricCollector) sendMultipleMetrics(metrics []MetricRequest) error {
	url := fmt.Sprintf("http://%s/updates/", ms.config.Address)

	chunks := rangeChunks(10, metrics)

	for i, chunk := range chunks {
		log.Printf("Sending chunk %d of %d, chunk size %d\n", i+1, len(chunks), len(chunk))
		msr := MetricsRequest{Metrics: chunk}

		data, err := json.Marshal(msr)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("response is not OK, status: %d", resp.StatusCode)
		}

		log.Printf("chunk %d sent successfully", i+1)
	}

	return nil
}

func rangeChunks(chunkSize int, metrics []MetricRequest) [][]MetricRequest {
	// const chunkSize = 3
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
		ms.sendMultipleMetrics(metrics) //
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
