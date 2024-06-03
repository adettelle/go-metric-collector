package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	store "github.com/adettelle/go-metric-collector/internal/storage"
)

type MetricApi struct {
	Storage store.Storage
}

func NewMetricApi(storage store.Storage) *MetricApi {
	return &MetricApi{
		Storage: storage,
	}
}

// CreateMetric adds metric into MemStorage
// http://localhost:8080/update/counter/someMetric/527
func (metricApi *MetricApi) CreateMetric(w http.ResponseWriter, r *http.Request) {
	// log.Println("Req:_________________", r)
	metricNameToSearch := r.PathValue("metric_name")
	metricValueToSearch := r.PathValue("metric_value")
	metricTypeToSearch := r.PathValue("metric_type")
	// log.Println("!!!!!!", metricNameToSearch, metricValueToSearch, metricTypeToSearch)

	switch {
	case metricTypeToSearch == "gauge":
		value, err := strconv.ParseFloat(metricValueToSearch, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metricApi.Storage.AddGaugeMetric(metricNameToSearch, value)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Created"))

	case metricTypeToSearch == "counter":
		value, err := strconv.ParseInt(metricValueToSearch, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metricApi.Storage.AddCounterMetric(metricNameToSearch, value)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Created"))
		fmt.Println(metricApi.Storage) // {map[] map[someMetric:[527]]}

	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No such metric"))
		return
	}
}
