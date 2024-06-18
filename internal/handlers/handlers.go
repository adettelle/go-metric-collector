package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// интерфейс для взаимодействия с хранилищем MemStorage и другими хранилищами, напрмер, fileStorage
type StorageInterfacer interface {
	GetGaugeMetric(name string) (float64, bool)
	GetCounterMetric(name string) (int64, bool)
	WriteMetricsReport(w io.Writer)
	AddGaugeMetric(name string, value float64)
	AddCounterMetric(name string, value int64)
}

type MetricAPI struct {
	Storage StorageInterfacer
}

func NewMetricAPI(storage StorageInterfacer) *MetricAPI {
	return &MetricAPI{
		Storage: storage,
	}
}

// CreateMetric adds metric into MemStorage
// POST http://localhost:8080/update/counter/someMetric/527
func (ma *MetricAPI) CreateMetric(w http.ResponseWriter, r *http.Request) {
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
		ma.Storage.AddGaugeMetric(metricName, value)

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
		ma.Storage.AddCounterMetric(metricName, value)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("Created"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// fmt.Println(ma.Storage) // {map[] map[someMetric:[527]]}

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
func (ma *MetricAPI) GetMetricByValue(w http.ResponseWriter, r *http.Request) {
	metricNameToSearch := r.PathValue("metric_name")
	metricTypeToSearch := r.PathValue("metric_type")
	switch {
	case metricTypeToSearch == "counter":
		metric, metricExists := ma.Storage.GetCounterMetric(metricNameToSearch)
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
		metric, metricExists := ma.Storage.GetGaugeMetric(metricNameToSearch)
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

func (ma *MetricAPI) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	ma.Storage.WriteMetricsReport(w)
}
