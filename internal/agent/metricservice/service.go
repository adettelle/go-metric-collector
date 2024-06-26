package metricservice

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	store "github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

func sendMetric(addr string, metricType string, name string, value float64) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", addr, metricType, name, value)
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

func sendAllMetrics(ms *store.MemStorage, addr string) error {
	for name, value := range ms.Gauge {
		err := sendMetric(addr, "gauge", name, value)
		if err != nil {
			log.Printf("Couldn't send metric, %s", err.Error())

		} else {
			log.Printf("Metric sent %v: %v", name, value)
		}
	}

	for name, value := range ms.Counter {
		err := sendMetric(addr, "counter", name, float64(value))
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
func SendLoop(delay time.Duration, metricsStorage *store.MemStorage, addr string, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for range ticker.C {
		log.Println("Sending metrics")
		err := sendAllMetrics(metricsStorage, addr)
		if err != nil {
			log.Fatal(err)
		}
		metricsStorage.Reset()
	}
}

// retrieveLoop gets all metrics from MemStorage to the server with delay
func RetrieveLoop(delay time.Duration, metricsStorage *store.MemStorage, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for range ticker.C {
		log.Println("Retrieving metrics")
		retrieveAllMetrics(metricsStorage)
	}
}

// retrieveAllMetrics получает все метрики из пакета runtime
// и собирает дополнительные метрики (PollCount и RandomValue)
func retrieveAllMetrics(metricsStorage *store.MemStorage) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metricsStorage.AddCounterMetric("PollCount", 1)

	metricsStorage.AddGaugeMetric("RandomValue", rand.Float64())

	metricsStorage.AddGaugeMetric("Alloc", float64(m.Alloc))
	metricsStorage.AddGaugeMetric("BuckHashSys", float64(m.BuckHashSys))
	metricsStorage.AddGaugeMetric("Frees", float64(m.Frees))
	metricsStorage.AddGaugeMetric("GCCPUFraction", m.GCCPUFraction)
	metricsStorage.AddGaugeMetric("GCSys", float64(m.GCSys))
	metricsStorage.AddGaugeMetric("HeapAlloc", float64(m.HeapAlloc))
	metricsStorage.AddGaugeMetric("HeapIdle", float64(m.HeapIdle))
	metricsStorage.AddGaugeMetric("HeapInuse", float64(m.HeapInuse))
	metricsStorage.AddGaugeMetric("HeapObjects", float64(m.HeapObjects))
	metricsStorage.AddGaugeMetric("HeapReleased", float64(m.HeapReleased))
	metricsStorage.AddGaugeMetric("HeapSys", float64(m.HeapSys))
	metricsStorage.AddGaugeMetric("LastGC", float64(m.LastGC))
	metricsStorage.AddGaugeMetric("Lookups", float64(m.Lookups))
	metricsStorage.AddGaugeMetric("MCacheInuse", float64(m.MCacheInuse))
	metricsStorage.AddGaugeMetric("MCacheSys", float64(m.MCacheSys))
	metricsStorage.AddGaugeMetric("MSpanInuse", float64(m.MSpanInuse))
	metricsStorage.AddGaugeMetric("MSpanSys", float64(m.MSpanSys))
	metricsStorage.AddGaugeMetric("Mallocs", float64(m.Mallocs))
	metricsStorage.AddGaugeMetric("NextGC", float64(m.NextGC))
	metricsStorage.AddGaugeMetric("NumForcedGC", float64(m.NumForcedGC))
	metricsStorage.AddGaugeMetric("NumGC", float64(m.NumGC))
	metricsStorage.AddGaugeMetric("OtherSys", float64(m.OtherSys))
	metricsStorage.AddGaugeMetric("PauseTotalNs", float64(m.PauseTotalNs))
	metricsStorage.AddGaugeMetric("StackInuse", float64(m.StackInuse))
	metricsStorage.AddGaugeMetric("StackSys", float64(m.StackSys))
	metricsStorage.AddGaugeMetric("Sys", float64(m.Sys))
	metricsStorage.AddGaugeMetric("TotalAlloc", float64(m.TotalAlloc))
}
