// The web controller layer is responsible for handling incoming http requests.
package api

import (
	"bytes"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/adettelle/go-metric-collector/internal/db"
	"github.com/adettelle/go-metric-collector/internal/security"
	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/server/service"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

// Storager defines an interface for interacting with various storage mechanisms,
// such as MemStorage or fileStorage. It describes operations to store, retrieve,
// check for existence, and delete a "metric" entity.
type Storager interface {
	GetGaugeMetric(name string) (float64, bool, error)
	GetCounterMetric(name string) (int64, bool, error)
	AddGaugeMetric(name string, value float64) error
	AddCounterMetric(name string, value int64) error
	GetAllGaugeMetrics() (map[string]float64, error)
	GetAllCounterMetrics() (map[string]int64, error)
	Finalize() error // отрабатывает завершение приложения (при штатном завершении работы)
}

// MetricHandlers contains dependencies for handling HTTP requests related
// to metrics, including a storage mechanism and configuration.
type MetricHandlers struct {
	Storager Storager
	Config   *config.Config
	DBCon    db.DBConnector // new
}

func NewMetricHandlers(storager Storager, config *config.Config) *MetricHandlers {
	return &MetricHandlers{
		Storager: storager,
		Config:   config,
		DBCon:    db.NewDBConnection(config.DBParams),
	}
}

// MetricUpdate handles HTTP requests to update a metric,
// accepting a JSON object in the request body.
func (mh *MetricHandlers) MetricUpdate(w http.ResponseWriter, r *http.Request) {
	var err error
	var ok bool

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var metric memstorage.Metric
	var buf bytes.Buffer

	// Read the request body
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Deserialize JSON into Metric
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch {
	case metric.MType == "gauge":
		if err = mh.Storager.AddGaugeMetric(metric.ID, *metric.Value); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		_, ok, err = mh.Storager.GetGaugeMetric(metric.ID)
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
		err = mh.Storager.AddCounterMetric(metric.ID, *metric.Delta)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		_, ok, err = mh.Storager.GetCounterMetric(metric.ID)
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
		_, err = w.Write([]byte("No such metric"))
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

// MetricValue retrieves the current value of a metric provided in the
// request body as a JSON object.
func (mh *MetricHandlers) MetricValue(w http.ResponseWriter, r *http.Request) {
	var err error

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var metric memstorage.Metric
	var buf bytes.Buffer

	// Read the request body
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Deserialize JSON into Metric
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch {
	case metric.MType == "gauge":
		value, ok, gaugeMetricErr := mh.Storager.GetGaugeMetric(metric.ID)
		if gaugeMetricErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metric.Value = &value

	case metric.MType == "counter":
		value, ok, counterMetricErr := mh.Storager.GetCounterMetric(metric.ID)
		if counterMetricErr != nil {
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
		_, err = w.Write([]byte("No such metric"))
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

// CreateMetric adds a new metric with a specific name and value into MemStorage
// POST http://localhost:8080/update/counter/someMetric/527
func (mh *MetricHandlers) CreateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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
		err = mh.Storager.AddGaugeMetric(metricName, value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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
		err = mh.Storager.AddCounterMetric(metricName, value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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

// GetMetric retrieves the value of a metric with the specified type
// and name from MemStorage
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
	_, err := mh.DBCon.Connect() // db.ConnectWithRerties(mh.Config.DBParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// MetricsUpdate processes an HTTP POST request that contains a batch of
// metrics ([]Metrics) in JSON format.
func (mh *MetricHandlers) MetricsUpdate(w http.ResponseWriter, r *http.Request) {
	var err error
	var ok bool

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var Metrics []memstorage.Metric
	var buf bytes.Buffer

	// читаем тело запроса
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// log.Println("mh.Config.Key:", mh.Config.Key)
	if mh.Config.Key != "" {
		// вычисляем хеш и сравниваем в HTTP-заголовке запроса с именем HashSHA256
		hash := security.CreateSign(buf.String(), mh.Config.Key)
		if !hmac.Equal([]byte(hash), []byte(r.Header.Get("HashSHA256"))) {
			log.Println("The signature is incorrect")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Println("The signature is authentic")
	}

	// десериализуем JSON в Metrric
	if err = json.Unmarshal(buf.Bytes(), &Metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, metric := range Metrics {
		switch {
		case metric.MType == "gauge":
			if err = mh.Storager.AddGaugeMetric(metric.ID, *metric.Value); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")

			_, ok, err = mh.Storager.GetGaugeMetric(metric.ID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

		case metric.MType == "counter":
			if err = mh.Storager.AddCounterMetric(metric.ID, *metric.Delta); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")

			_, ok, err = mh.Storager.GetCounterMetric(metric.ID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}

		default:
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte("No such metric"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
	}

	resp := []byte("{\"result\": \"ok\"}")
	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
