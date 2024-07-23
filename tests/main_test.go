package tests

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adettelle/go-metric-collector/internal/api"
	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
	"github.com/stretchr/testify/assert"
)

type metric struct {
	metricType  string
	metricName  string
	metricValue string
}

var metrics = []metric{
	{
		metricType:  "counter",
		metricName:  "C1",
		metricValue: "123",
	},
	{
		metricType:  "counter",
		metricName:  "C1",
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

var Config *config.Config = &config.Config{
	Address:       "localhost:8080",
	StoreInterval: 5,
	StoragePath:   "./test.json",
	Restore:       false,
}

// var mStorage = storage.NewMemStorage()

func TestAddCounterMetric(t *testing.T) {
	ms, err := memstorage.New(false, "", 0)
	if err != nil {
		log.Fatal(err)
	}

	name := "someMetric"
	var value int64 = 525

	// записали метрику в хранилище
	lenBeforeAdding := len(ms.Counter)
	ms.AddCounterMetric(name, value)
	lenAfterAdding := len(ms.Counter)
	assert.NotEqual(t, lenBeforeAdding, lenAfterAdding)

	// проверка наличия метрики в map
	val1, ok, err := ms.GetCounterMetric(name)
	assert.Equal(t, value, val1)
	assert.True(t, ok)
	assert.Nil(t, err)

	// проверка добавление уже сущ-ей метрики
	ms.AddCounterMetric(name, value)
	val2, ok, err := ms.GetCounterMetric(name)
	assert.Equal(t, int64(1050), val2)
	assert.True(t, ok)
	assert.Nil(t, err)

	// проверка получения несущ-ей метрики
	unrealName := "UnrealMetric"
	var zero int64 = 0
	v, ok, err := ms.GetCounterMetric(unrealName)
	assert.False(t, ok)
	assert.Equal(t, v, zero)
	assert.Nil(t, err)
}

func TestAddGaugeMetric(t *testing.T) {
	ms, err := memstorage.New(false, "", 0)
	if err != nil {
		log.Fatal(err)
	}

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
	val1, ok, err := ms.GetGaugeMetric(name)
	assert.Equal(t, value, val1)
	assert.True(t, ok)
	assert.Nil(t, err)

	// проверка добавление уже сущ-ей метрики
	ms.AddGaugeMetric(name, value)
	val2, ok, err := ms.GetGaugeMetric(name)
	assert.Equal(t, value, val2) // при добавлении сущ-ей метрики метрика заменятеся на новую
	assert.True(t, ok)
	assert.Nil(t, err)

	// проверка получения несущ-ей метрики
	unrealName := "UnrealMetric"
	var zero float64 = 0
	v, ok, err := ms.GetGaugeMetric(unrealName)
	assert.False(t, ok)
	assert.Equal(t, v, zero)
	assert.Nil(t, err)
}

// 200
func TestPostCounterMetric(t *testing.T) {
	m0 := metrics[0]
	query := "/update/" + m0.metricType + "/" + m0.metricName + "/" + m0.metricValue
	// query := "/update/counter/counterMetric/525" // http://localhost:8080
	request := httptest.NewRequest(http.MethodPost, query, nil)

	request.SetPathValue("metric_type", m0.metricType)
	request.SetPathValue("metric_name", m0.metricName)
	request.SetPathValue("metric_value", m0.metricValue)

	res := testPostMetric(t, request, http.StatusOK, "Created")
	defer res.Body.Close()
}

func testPostMetric(t *testing.T, request *http.Request, expectedStatus int, expectedBody string) *http.Response {

	metricStore, err := memstorage.New(false, "", 0)
	if err != nil {
		log.Fatal(err)
	}
	mAPI := api.NewMetricHandlers(metricStore, Config)
	w := httptest.NewRecorder()
	mAPI.CreateMetric(w, request)

	res := w.Result()
	assert.Equal(t, expectedStatus, res.StatusCode)

	bodyStr, _ := io.ReadAll(res.Body)
	fmt.Println("Body:", string(bodyStr))
	assert.Equal(t, expectedBody, string(bodyStr))

	return res
}

const tmpl = `
<html>

	<body>
		<h1>Gauge metrics</h1>
    	<table> 

     		<tr>
				<td>G1</td>
				<td>123</td> 
			</tr>

			<tr>
				<td>G2</td>
				<td>150984.573</td> 
			</tr>

		</table>

		<h1>Counter metrics</h1>
    	<table> 

     		<tr>
				<td>C1</td>
				<td>579</td> 
			</tr>

		</table>
	</body>

</html>
	`

func TestGetAllMetrics(t *testing.T) {

	metricStore, err := memstorage.New(false, "", 0)
	if err != nil {
		log.Fatal(err)
	}
	metricAPI := api.NewMetricHandlers(metricStore, Config)
	metricStore.AddCounterMetric("C1", 123)
	metricStore.AddCounterMetric("C1", 456)
	metricStore.AddGaugeMetric("G1", 123)
	metricStore.AddGaugeMetric("G2", 150984.573)

	query := "/"
	w := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodGet, query, nil)
	metricAPI.GetAllMetrics(w, request)

	expectedBody := strings.Join(strings.Fields(tmpl), "")
	body := strings.Join(strings.Fields(w.Body.String()), "")
	assert.Equal(t, expectedBody, body)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

}

// http://localhost:8080/value/counter/HeapAlloc
func testGetValue(mType, mName string, mAPI *api.MetricHandlers) (string, int) {

	w := httptest.NewRecorder()
	//defer w.Result().Body.Close()

	path := fmt.Sprintf("/value/%s/%s", mType, mName)
	req := httptest.NewRequest(http.MethodGet, path, nil)

	req.SetPathValue("metric_type", mType)
	req.SetPathValue("metric_name", mName)

	mAPI.GetMetricByValue(w, req)

	res := w.Result()
	defer res.Body.Close()
	code := res.StatusCode

	return w.Body.String(), code
}
func TestGetMetricByValue(t *testing.T) {

	metricStore, err := memstorage.New(false, "", 0)
	if err != nil {
		log.Fatal(err)
	}
	mAPI := api.NewMetricHandlers(metricStore, Config)

	metricStore.AddCounterMetric("C1", 123)
	metricStore.AddCounterMetric("C1", 456)
	metricStore.AddGaugeMetric("G1", 123)
	metricStore.AddGaugeMetric("G2", 150984.573)

	var testTable = []struct {
		mType  string
		mName  string
		want   string
		status int
	}{
		{"gauge", "G1", "123", http.StatusOK},
		{"gauge", "G2", "150984.573", http.StatusOK},
		{"counter", "C1", "579", http.StatusOK},
		// проверим на ошибочный запрос
		{"count", "a6", "No such metric type", http.StatusNotFound},
	}

	for _, v := range testTable {
		resp, statusCode := testGetValue(v.mType, v.mName, mAPI)
		assert.Equal(t, v.status, statusCode)
		assert.Equal(t, v.want, resp)
	}
}
