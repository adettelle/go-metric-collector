// слой веб контроллер отвечает за обработку входящих http запросов
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/server/service"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

// интерфейс для взаимодействия с хранилищем MemStorage и другими хранилищами, напрмер, fileStorage
// type Storager interface {
// 	GetGaugeMetric(name string) (float64, bool)
// 	GetCounterMetric(name string) (int64, bool)
// 	AddGaugeMetric(name string, value float64)
// 	AddCounterMetric(name string, value int64)
// 	GetAllGaugeMetrics() map[string]float64
// 	GetAllCounterMetrics() map[string]int64
// }

func (mh *MetricHandlers) JSONHandlerUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var metric memstorage.Metric
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// десериализуем JSON в Metrric
	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch {
	case metric.MType == "gauge":
		mh.Storage.AddGaugeMetric(metric.ID, *metric.Value)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, ok := mh.Storage.GetGaugeMetric(metric.ID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

	case metric.MType == "counter":
		mh.Storage.AddCounterMetric(metric.ID, *metric.Delta)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, ok := mh.Storage.GetCounterMetric(metric.ID)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("No such metric"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	resp, err := json.Marshal(metric)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (mh *MetricHandlers) JSONHandlerValue(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var metric memstorage.Metric
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// десериализуем JSON в Metrric
	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch {
	case metric.MType == "gauge":
		mh.Storage.GetGaugeMetric(metric.ID)

		value := mh.Storage.Gauge[metric.ID]
		metric.Value = &value

	case metric.MType == "counter":
		mh.Storage.GetCounterMetric(metric.ID)

		value := mh.Storage.Counter[metric.ID]
		metric.Delta = &value

	default:
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("No such metric"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	resp, err := json.Marshal(metric)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

type MetricHandlers struct { // было MetricAPI
	Storage *memstorage.MemStorage // Storager
	Config  *config.Config
}

func NewMetricHandlers(storage *memstorage.MemStorage, config *config.Config) *MetricHandlers { //Storager // ранее был NewMetricAPI
	return &MetricHandlers{
		Storage: storage,
		Config:  config,
	}
}

// CreateMetric adds metric into MemStorage
// POST http://localhost:8080/update/counter/someMetric/527
func (mh *MetricHandlers) CreateMetric(w http.ResponseWriter, r *http.Request) {
	metricName := r.PathValue("metric_name")
	metricValue := r.PathValue("metric_value")
	metricType := r.PathValue("metric_type")

	switch {
	case metricType == "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mh.Storage.AddGaugeMetric(metricName, value)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("Created"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case metricType == "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mh.Storage.AddCounterMetric(metricName, value)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("Created"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("No such metric"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
}

// GetMetric gets metric from MemStorage
// GET http://localhost:8080/value/counter/HeapAlloc
func (mh *MetricHandlers) GetMetricByValue(w http.ResponseWriter, r *http.Request) {
	metricNameToSearch := r.PathValue("metric_name")
	metricTypeToSearch := r.PathValue("metric_type")
	switch {
	case metricTypeToSearch == "counter":
		metric, metricExists := mh.Storage.GetCounterMetric(metricNameToSearch)
		if !metricExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err := w.Write([]byte(fmt.Sprintf("%v", metric)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case metricTypeToSearch == "gauge":
		metric, metricExists := mh.Storage.GetGaugeMetric(metricNameToSearch)
		if !metricExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err := w.Write([]byte(fmt.Sprintf("%v", metric)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("No such metric type"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

}

func (mh *MetricHandlers) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	service.WriteMetricsReport(mh.Storage, w)
	w.WriteHeader(http.StatusOK)
}
