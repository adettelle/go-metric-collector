package tests

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adettelle/go-metric-collector/internal/handlers"
	"github.com/adettelle/go-metric-collector/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestAddCounterMetric(t *testing.T) {
	ms := storage.NewMemStorage()

	name := "someMetric"
	var value int64 = 525

	// записали метрику в хранилище
	lenBeforeAdding := len(ms.Counter)
	ms.AddCounterMetric(name, value)
	lenAfterAdding := len(ms.Counter)
	assert.NotEqual(t, lenBeforeAdding, lenAfterAdding)

	// проверка наличия метрики в map
	val1, err := ms.GetCounterMetric(name)
	assert.Equal(t, value, val1)
	assert.True(t, err)

	// проверка добавление уже сущ-ей метрики
	ms.AddCounterMetric(name, value)
	val2, err := ms.GetCounterMetric(name)
	assert.Equal(t, int64(1050), val2)
	assert.True(t, err)

	// проверка получения несущ-ей метрики
	unrealName := "UnrealMetric"
	var zero int64 = 0
	v, err := ms.GetCounterMetric(unrealName)
	assert.False(t, err)
	assert.Equal(t, v, zero)
}

func TestAddGaugeMetric(t *testing.T) {
	ms := storage.NewMemStorage()

	name := "someMetric"
	var value float64 = 527

	// записали метрику в хранилище
	lenBeforeAdding := len(ms.Gauge)
	ms.AddGaugeMetric(name, value)
	lenAfterAdding := len(ms.Gauge)

	assert.NotEqual(t, lenBeforeAdding, lenAfterAdding)
	checkValue, ok := ms.Gauge[name]
	if ok {
		assert.Equal(t, value, checkValue)
	}

	// проверка наличия метрики в map
	val1, err := ms.GetGaugeMetric(name)
	assert.Equal(t, value, val1)
	assert.True(t, err)

	// проверка добавление уже сущ-ей метрики
	ms.AddGaugeMetric(name, value)
	val2, err := ms.GetGaugeMetric(name)
	assert.Equal(t, value, val2) // при добавлении сущ-ей метрики метрика заменятеся на новую
	assert.True(t, err)

	// проверка получения несущ-ей метрики
	unrealName := "UnrealMetric"
	var zero float64 = 0
	v, err := ms.GetGaugeMetric(unrealName)
	assert.False(t, err)
	assert.Equal(t, v, zero)
}

// 200
func TestPostCounterMetric(t *testing.T) {
	// metrics := map[string]string{

	// }
	type metric struct {
		metricType  string
		metricName  string
		metricValue string
	}
	metrics := []metric{
		{
			metricType:  "counter",
			metricName:  "C1",
			metricValue: "123",
		},
		{
			metricType:  "counter",
			metricName:  "C2",
			metricValue: "567",
		},
		{
			metricType:  "gauge",
			metricName:  "G1",
			metricValue: "123",
		},
		{
			metricType:  "gauge",
			metricName:  "G2",
			metricValue: "456",
		},
	}
	m0 := metrics[0]
	query := "/update/" + m0.metricType + "/" + m0.metricName + "/" + m0.metricValue
	// query := "/update/counter/counterMetric/525" // http://localhost:8080
	request := httptest.NewRequest(http.MethodPost, query, nil)

	// request.SetPathValue("metricType", "counter")
	// request.SetPathValue("metricName", "counterMetric")
	// request.SetPathValue("metricValue", "525")

	request.SetPathValue("metric_type", m0.metricType)
	request.SetPathValue("metric_name", m0.metricName)
	request.SetPathValue("metric_value", m0.metricValue)

	res := testPostMetric(t, request, http.StatusOK, "Created")
	defer res.Body.Close()
}

func testPostMetric(t *testing.T, request *http.Request, expectedStatus int, expectedBody string) *http.Response {
	metricStore := storage.NewMemStorage()
	mAPI := handlers.NewMetricApi(metricStore)
	w := httptest.NewRecorder()
	mAPI.CreateMetric(w, request)

	res := w.Result()
	defer res.Body.Close()
	assert.Equal(t, expectedStatus, res.StatusCode)

	bodyStr, _ := io.ReadAll(res.Body)
	fmt.Println("Body:", string(bodyStr))
	assert.Equal(t, expectedBody, string(bodyStr))

	return res
}
