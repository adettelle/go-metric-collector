package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	store "github.com/adettelle/go-metric-collector/internal/storage"
)

// Поскольку Storage - это интерфейс, то ссылки на него быть не должно!!!!
// А когда был MemStorage, то была ссылка: Storage *store.Storage
type MetricApi struct {
	Storage store.Storage
}

// Поскольку Storage - это интерфейс, то ссылки на него быть не должно!!!!
// А когда был MEmStorage, то была ссылка: storage *store.MemStorage и
func NewMetricApi(storage store.Storage) *MetricApi {
	return &MetricApi{
		Storage: storage,
	}
}

// CreateMetric adds metric into MemStorage
// http://localhost:8080/update/counter/someMetric/527
func (metricApi *MetricApi) CreateMetric(w http.ResponseWriter, r *http.Request) {
	log.Println("Req:_________________", r)
	metricNameToSearch := r.PathValue("metric_name")
	metricValueToSearch := r.PathValue("metric_value")
	metricTypeToSearch := r.PathValue("metric_type")
	log.Println("!!!!!!", metricNameToSearch, metricValueToSearch, metricTypeToSearch)

	switch {
	case metricTypeToSearch == "gauge":
		value, err := strconv.ParseFloat(metricValueToSearch, 32)
		if err != nil {
			log.Fatal(err)
		}
		metricApi.Storage.AddGaugeMetric(metricNameToSearch, value)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Created"))

	case metricTypeToSearch == "counter":
		value, err := strconv.ParseInt(metricValueToSearch, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		metricApi.Storage.AddCounterMetric(metricNameToSearch, value)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Created"))
		fmt.Println(metricApi.Storage) // {map[] map[someMetric:[527]]}
		// что это такое в начале стрки????
		// &{{{0 0} 0 0 {{} 0} {{} 0}} map[Alloc:827680 BuckHashSys:6919...] map[PollCount:10]}

	case metricTypeToSearch == "":
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Metric name is empty"))
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No such metric"))
	}
}
