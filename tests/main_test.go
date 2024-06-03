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
		metric_type  string
		metric_name  string
		metric_value string
	}
	metrics := []metric{
		{
			metric_type:  "counter",
			metric_name:  "C1",
			metric_value: "123",
		},
		{
			metric_type:  "counter",
			metric_name:  "C2",
			metric_value: "567",
		},
		{
			metric_type:  "gauge",
			metric_name:  "G1",
			metric_value: "123",
		},
		{
			metric_type:  "gauge",
			metric_name:  "G2",
			metric_value: "456",
		},
	}
	m1 := metrics[0]
	query := "/update/" + m1.metric_type + "/" + m1.metric_name + m1.metric_value
	// query := "/update/counter/counterMetric/525" // http://localhost:8080
	request := httptest.NewRequest(http.MethodPost, query, nil)

	// request.SetPathValue("metric_type", "counter")
	// request.SetPathValue("metric_name", "counterMetric")
	// request.SetPathValue("metric_value", "525")

	request.SetPathValue("metric_type", m1.metric_type)
	request.SetPathValue("metric_name", m1.metric_name)
	request.SetPathValue("metric_value", m1.metric_value)

	testPostMetric(t, request, http.StatusOK, "Created")

}

func testPostMetric(t *testing.T, request *http.Request, expectedStatus int, expectedBody string) *http.Response {
	metricStore := storage.NewMemStorage()
	mApi := handlers.NewMetricApi(metricStore)

	w := httptest.NewRecorder()
	mApi.CreateMetric(w, request)

	res := w.Result()
	assert.Equal(t, expectedStatus, res.StatusCode)

	bodyStr, _ := io.ReadAll(res.Body)
	fmt.Println("Body:", string(bodyStr))
	assert.Equal(t, expectedBody, string(bodyStr))

	return res
}
