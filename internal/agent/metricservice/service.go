// Сервисный слой отвечает за сбор и отправку метрик на удаленный сервер
package metricservice

import (
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

// Структура MetricService получает и рассылает метрики, запускает свои циклы (Loop)
type MetricService struct {
	config *config.Config
	// store         StorageInterfase
	metricStorage *mstore.MemStorage
}

// type StorageInterfase interface {
// }

func NewMetricService(config *config.Config, metricStorage *mstore.MemStorage) *MetricService { // store StorageInterfase,

	return &MetricService{
		config: config,
		// store:         store,
		metricStorage: metricStorage,
	}
}

func (ms *MetricService) sendMetric(metricType string, name string, value float64) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", ms.config.Address, metricType, name, value)
	req, err := http.NewRequest(http.MethodPost, url, nil)
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
	return nil
}

func (ms *MetricService) sendAllMetrics() error {
	for name, value := range ms.metricStorage.Gauge {
		err := ms.sendMetric("gauge", name, value)
		if err != nil {
			log.Printf("Couldn't send metric, %s", err.Error())

		} else {
			log.Printf("Metric sent %v: %v", name, value)
		}
	}

	for name, value := range ms.metricStorage.Counter {
		err := ms.sendMetric("counter", name, float64(value))
		if err != nil {
			log.Printf("Couldn't send metric, %s", err.Error())
			return err
		} else {
			log.Printf("Metric sent %v: %v", name, value)
		}
	}

	return nil
}

// sendLoop sends all metrics to the server (MemStorage) with delay
func (ms *MetricService) SendLoop(delay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for range ticker.C {
		log.Println("Sending metrics")
		err := ms.sendAllMetrics()
		if err != nil {
			log.Fatal(err)
		}
		ms.metricStorage.Reset()
	}
}

// retrieveLoop gets all metrics from MemStorage to the server with delay
func (ms *MetricService) RetrieveLoop(delay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for range ticker.C {
		log.Println("Retrieving metrics")
		ms.retrieveAllMetrics()
	}
}

// retrieveAllMetrics получает все метрики из пакета runtime
// и собирает дополнительные метрики (PollCount и RandomValue)
func (ms *MetricService) retrieveAllMetrics() {
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