package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	store "github.com/adettelle/go-metric-collector/internal/storage"
)

type MetricAPI struct {
	Storage store.Storage
}

func NewMetricAPI(storage store.Storage) *MetricAPI {
	return &MetricAPI{
		Storage: storage,
	}
}

// CreateMetric adds metric into MemStorage
// POST http://localhost:8080/update/counter/someMetric/527
func (MetricAPI *MetricAPI) CreateMetric(w http.ResponseWriter, r *http.Request) {
	metricName := r.PathValue("metric_name")
	metricValue := r.PathValue("metric_value")
	metricType := r.PathValue("metric_type")

	switch {
	case metricType == "gauge":
		value, err := strconv.ParseFloat(metricValue, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		MetricAPI.Storage.AddGaugeMetric(metricName, value)

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
		MetricAPI.Storage.AddCounterMetric(metricName, value)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("Created"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// fmt.Println(MetricAPI.Storage) // {map[] map[someMetric:[527]]}

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
func (mAPI *MetricAPI) GetMetricByValue(w http.ResponseWriter, r *http.Request) {
	metricNameToSearch := r.PathValue("metric_name")
	metricTypeToSearch := r.PathValue("metric_type")

	switch {
	case metricTypeToSearch == "counter":
		metric, metricExists := mAPI.Storage.GetCounterMetric(metricNameToSearch)
		if !metricExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err := w.Write([]byte(fmt.Sprintf("%v: %v", metricNameToSearch, metric)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case metricTypeToSearch == "gauge":
		metric, metricExists := mAPI.Storage.GetGaugeMetric(metricNameToSearch)
		if !metricExists {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err := w.Write([]byte(fmt.Sprintf("%v: %v", metricNameToSearch, metric)))
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

func (mAPI *MetricAPI) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	mAPI.Storage.GetAllMetric(w)
}
