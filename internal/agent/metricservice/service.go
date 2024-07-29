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
	"github.com/adettelle/go-metric-collector/internal/security"
	"github.com/adettelle/go-metric-collector/pkg/retries"

	m "github.com/adettelle/go-metric-collector/internal/agent/metrics"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Структура MetricCollector получает и рассылает метрики, запускает свои циклы (Loop)
type MetricService struct { // MetricCollector
	// config *config.Config // был нужен только для генерации url
	// store         StorageInterfase
	metricAccumulator *m.MetricAccumulator // *metrics.MetricAccumulator
	client            *http.Client
	url               string
	maxRequestRetries int
	encryptionKey     string
	rateLimit         int
}

func NewMetricService(config *config.Config, metricAccumulator *m.MetricAccumulator, client *http.Client) *MetricService { // store StorageInterfase,

	return &MetricService{
		// config: config,
		// store:         store,
		metricAccumulator: metricAccumulator,
		client:            client,
		url:               fmt.Sprintf("http://%s/updates/", config.Address),
		maxRequestRetries: config.MaxRequestRetries,
		encryptionKey:     config.Key,
		rateLimit:         config.RateLimit,
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

// type MetricsRequest []MetricRequest

func (ms *MetricService) sendMultipleMetrics(metrics []MetricRequest,
	workerRequests chan<- []MetricRequest) error {
	// url := fmt.Sprintf("http://%s/updates/", ms.config.Address)

	chunks := rangeChunks(10, metrics)

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

func (ms *MetricService) doSend(data *bytes.Buffer) error {
	req, err := http.NewRequest(http.MethodPost, ms.url, data)
	if err != nil {
		return err
	}

	if ms.encryptionKey != "" {
		// вычисляем хеш и передаем в HTTP-заголовке запроса с именем HashSHA256
		hash := security.CreateSign(data.String(), ms.encryptionKey)
		log.Println(data.String(), string(hash))
		req.Header.Set("HashSHA256", string(hash))
	}

	resp, err := ms.client.Do(req) // http.DefaultClient.Do(req)
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
		// log.Printf("Sending chunk %d of %d, chunk size %d\n", i+1, len(chunks), len(chunk))
		log.Printf("Sending chunk on worker %d\n", id)

		data, err := json.Marshal(chunk)
		if err != nil {
			// return err
			log.Printf("error %v in sending chun in worker %d\n", err, id)
			results <- false
			continue // прерываем итерацию, но не сам worker и не цикл
		}

		_, err = retries.RunWithRetries("Send metrics request",
			ms.maxRequestRetries,
			func() (*any, error) {
				err := ms.doSend(bytes.NewBuffer(data))
				return nil, err
			}, isRetriableError)

		if err != nil {
			log.Printf("error %v in sending chun in worker %d\n", err, id)
			results <- false
			continue // worker
		}

		log.Printf("chunk in worker %d sent successfully\n", id)
		results <- true
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

// AdditionalRetrieveLoop gets aditional metrics from MemStorage to the server with delay
func (ms *MetricService) AdditionalRetrieveLoop(delay time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second * delay)

	for range ticker.C {
		log.Println("Retrieving additional metrics")
		ms.retrieveAdditionalGaugeMetrics()
	}
}

// retrieveAllMetrics получает все метрики из пакета runtime
// и собирает дополнительные метрики (PollCount и RandomValue)
func (ms *MetricService) retrieveAllMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ms.metricAccumulator.AddCounterMetric("PollCount", 1)

	ms.metricAccumulator.AddGaugeMetric("RandomValue", rand.Float64())

	ms.metricAccumulator.AddGaugeMetric("Alloc", float64(m.Alloc))
	ms.metricAccumulator.AddGaugeMetric("BuckHashSys", float64(m.BuckHashSys))
	ms.metricAccumulator.AddGaugeMetric("Frees", float64(m.Frees))
	ms.metricAccumulator.AddGaugeMetric("GCCPUFraction", m.GCCPUFraction)
	ms.metricAccumulator.AddGaugeMetric("GCSys", float64(m.GCSys))
	ms.metricAccumulator.AddGaugeMetric("HeapAlloc", float64(m.HeapAlloc))
	ms.metricAccumulator.AddGaugeMetric("HeapIdle", float64(m.HeapIdle))
	ms.metricAccumulator.AddGaugeMetric("HeapInuse", float64(m.HeapInuse))
	ms.metricAccumulator.AddGaugeMetric("HeapObjects", float64(m.HeapObjects))
	ms.metricAccumulator.AddGaugeMetric("HeapReleased", float64(m.HeapReleased))
	ms.metricAccumulator.AddGaugeMetric("HeapSys", float64(m.HeapSys))
	ms.metricAccumulator.AddGaugeMetric("LastGC", float64(m.LastGC))
	ms.metricAccumulator.AddGaugeMetric("Lookups", float64(m.Lookups))
	ms.metricAccumulator.AddGaugeMetric("MCacheInuse", float64(m.MCacheInuse))
	ms.metricAccumulator.AddGaugeMetric("MCacheSys", float64(m.MCacheSys))
	ms.metricAccumulator.AddGaugeMetric("MSpanInuse", float64(m.MSpanInuse))
	ms.metricAccumulator.AddGaugeMetric("MSpanSys", float64(m.MSpanSys))
	ms.metricAccumulator.AddGaugeMetric("Mallocs", float64(m.Mallocs))
	ms.metricAccumulator.AddGaugeMetric("NextGC", float64(m.NextGC))
	ms.metricAccumulator.AddGaugeMetric("NumForcedGC", float64(m.NumForcedGC))
	ms.metricAccumulator.AddGaugeMetric("NumGC", float64(m.NumGC))
	ms.metricAccumulator.AddGaugeMetric("OtherSys", float64(m.OtherSys))
	ms.metricAccumulator.AddGaugeMetric("PauseTotalNs", float64(m.PauseTotalNs))
	ms.metricAccumulator.AddGaugeMetric("StackInuse", float64(m.StackInuse))
	ms.metricAccumulator.AddGaugeMetric("StackSys", float64(m.StackSys))
	ms.metricAccumulator.AddGaugeMetric("Sys", float64(m.Sys))
	ms.metricAccumulator.AddGaugeMetric("TotalAlloc", float64(m.TotalAlloc))
}

// retrieveAdditionalGaugeMetrics получает дополнительные метрики из пакета gopsutil
func (ms *MetricService) retrieveAdditionalGaugeMetrics() {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
	}
	cpu.Info()
	CPUutilizations, err := cpu.Percent(0, true)
	if err != nil {
		log.Fatal(err)
	}
	ms.metricAccumulator.AddGaugeMetric("TotalMemory", float64(v.Total))
	ms.metricAccumulator.AddGaugeMetric("FreeMemory", float64(v.Free))
	for i, CPUutilization := range CPUutilizations {
		ms.metricAccumulator.AddGaugeMetric(fmt.Sprintf("CPUutilization%d", i+1), CPUutilization)
	}
}
