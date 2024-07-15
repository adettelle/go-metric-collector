// слой веб контроллер отвечает за обработку входящих http запросов
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/adettelle/go-metric-collector/internal/db"
	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/server/service"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

// интерфейс для взаимодействия с хранилищем MemStorage и другими хранилищами, напрмер, fileStorage
type Storager interface {
	GetGaugeMetric(name string) (float64, bool, error)
	GetCounterMetric(name string) (int64, bool, error)
	AddGaugeMetric(name string, value float64) error
	AddCounterMetric(name string, value int64) error
	GetAllGaugeMetrics() (map[string]float64, error)
	GetAllCounterMetrics() (map[string]int64, error)
}

type MetricHandlers struct { // было MetricAPI
	Storager Storager // *memstorage.MemStorage // Storager
	Config   *config.Config
}

func NewMetricHandlers(storager Storager, config *config.Config) *MetricHandlers { //был storage *memstorage.MemStorage  // ранее был NewMetricAPI
	return &MetricHandlers{
		Storager: storager,
		Config:   config,
	}
}

// принимает в теле запроса метрику в формате json
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
		mh.Storager.AddGaugeMetric(metric.ID, *metric.Value)
		w.Header().Set("Content-Type", "application/json")
		// w.WriteHeader(http.StatusOK)

		_, ok, err := mh.Storager.GetGaugeMetric(metric.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)

	case metric.MType == "counter":
		mh.Storager.AddCounterMetric(metric.ID, *metric.Delta)
		w.Header().Set("Content-Type", "application/json")
		// w.WriteHeader(http.StatusOK)

		_, ok, err := mh.Storager.GetCounterMetric(metric.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)

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
		value, ok, err := mh.Storager.GetGaugeMetric(metric.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metric.Value = &value

	case metric.MType == "counter":
		value, ok, err := mh.Storager.GetCounterMetric(metric.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

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
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
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
		mh.Storager.AddGaugeMetric(metricName, value)

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
		mh.Storager.AddCounterMetric(metricName, value)

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
		metric, metricExists, err := mh.Storager.GetCounterMetric(metricNameToSearch)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !metricExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err = w.Write([]byte(fmt.Sprintf("%v", metric)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case metricTypeToSearch == "gauge":
		metric, metricExists, err := mh.Storager.GetGaugeMetric(metricNameToSearch)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !metricExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err = w.Write([]byte(fmt.Sprintf("%v", metric)))
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

// используем интерфейс mh.Storager (Reporter), у кого есть GetAllCounterMetrics, GetAllGaugeMetrics
func (mh *MetricHandlers) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	err := service.WriteMetricsReport(mh.Storager, w) // было mh *MetricHandlers
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (mh *MetricHandlers) CheckConnectionToDB(w http.ResponseWriter, r *http.Request) {
	log.Println("Checking DB")
	_, err := db.Connect(mh.Config.DBParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type MetricsRequest []memstorage.Metric

// принимает в теле запроса множество метрик в формате: []Metrics (списка метрик) в виде json
func (mh *MetricHandlers) MetricsHandlerUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var Metrics MetricsRequest
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// десериализуем JSON в Metrric
	if err := json.Unmarshal(buf.Bytes(), &Metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, metric := range Metrics {
		switch {
		case metric.MType == "gauge":
			err := mh.Storager.AddGaugeMetric(metric.ID, *metric.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			// w.WriteHeader(http.StatusOK)

			_, ok, err := mh.Storager.GetGaugeMetric(metric.ID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			//w.WriteHeader(http.StatusOK)

		case metric.MType == "counter":
			err := mh.Storager.AddCounterMetric(metric.ID, *metric.Delta)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			// w.WriteHeader(http.StatusOK)

			_, ok, err := mh.Storager.GetCounterMetric(metric.ID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			// w.WriteHeader(http.StatusOK)

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
	// resp, err := json.Marshal(Metrics)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	resp := []byte("{\"result\": \"ok\"}")
	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
