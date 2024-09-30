package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/adettelle/go-metric-collector/internal/mocks"
	"github.com/adettelle/go-metric-collector/internal/security"
	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/server/service"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------- Хендлер: GET /value/{metric_type}/{metric_name}
// test case with no such counter metric
func TestGetMetricByValueWithNoMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
		Config:   nil,
	}

	metricType := "counter"
	metricName := "Counter"
	metricValue := int64(0)
	metricExists := false

	reqURL := fmt.Sprintf("/value/%s/%s", metricType, metricName)

	m.EXPECT().GetCounterMetric(metricName).Return(metricValue, metricExists, nil)

	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	request.SetPathValue("metric_type", metricType)
	request.SetPathValue("metric_name", metricName)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	mh.GetMetricByValue(response, request)

	require.Equal(t, http.StatusNotFound, response.Code)
}

// test case with counter metric
func TestGetMetricCounterByValue(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
		Config:   nil,
	}

	metricType := "counter"
	metricName := "PollCount"
	metricValue := int64(10)
	metricExists := true

	reqURL := fmt.Sprintf("/value/%s/%s", metricType, metricName)

	// пишем, что хотим получить от заглушки
	m.EXPECT().GetCounterMetric(metricName).Return(metricValue, metricExists, nil)

	// тестируем хэндлер r.Get("/value/{metric_type}/{metric_name}", mware.WithLogging(mh.GetMetricByValue))
	// 1. Создаем запрос для обработчика
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	// без такого задания типа и имени метрик путь не собирается и не распознается!!!
	request.SetPathValue("metric_type", metricType)
	request.SetPathValue("metric_name", metricName)
	require.NoError(t, err)

	// 2. Вызваем функцию NewRecorder() *ResponseRecorder, которая возвращает
	// переменную типа *httptest.ResponseRecorder. Она будет использоваться для получения ответа.
	response := httptest.NewRecorder()
	// 3. Вызваем проверяемый обработчик, которому передаются запрос и переменная для получения ответа
	mh.GetMetricByValue(response, request)
	// 4. Вызваем метод (rw *ResponseRecorder) Result() *http.Response, чтобы получить ответ типа *http.Response.
	result := response.Result()
	defer result.Body.Close()

	// 5. Проверяем параметры ответа с ожидаемыми значениями
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "10", response.Body.String())
}

// test case with gauge metric
func TestGetMetricGaugeByValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
		Config:   nil,
	}

	var testTable = []struct {
		mType   string
		mName   string
		want    float64
		status  int
		mExists bool
	}{
		{"gauge", "G1", 123.0, http.StatusOK, true},
		{"gauge", "G2", 150984.573, http.StatusOK, true},
	}

	for _, v := range testTable {
		reqURL := fmt.Sprintf("/value/%s/%s", v.mType, v.mName)

		m.EXPECT().GetGaugeMetric(v.mName).Return(v.want, v.mExists, nil)

		request, err := http.NewRequest(http.MethodGet, reqURL, nil)

		request.SetPathValue("metric_type", v.mType)
		request.SetPathValue("metric_name", v.mName)
		require.NoError(t, err)

		response := httptest.NewRecorder()
		mh.GetMetricByValue(response, request)
		result := response.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusOK, response.Code)
		n, err := strconv.ParseFloat(response.Body.String(), 64)
		require.NoError(t, err)
		require.Equal(t, v.want, n)
	}
}

// проверим на ошибочный запрос
func TestGetMetricByWrongValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
		Config:   nil,
	}

	var testTable = []struct {
		mType   string
		mName   string
		want    float64
		status  int
		mExists bool
	}{
		{"some", "a6", 0, http.StatusNotFound, false},
	}

	for _, v := range testTable {
		reqURL := fmt.Sprintf("/value/%s/%s", v.mType, v.mName)

		// тестируем хэндлер r.Get("/value/{metric_type}/{metric_name}", mware.WithLogging(mh.GetMetricByValue))
		request, err := http.NewRequest(http.MethodGet, reqURL, nil)
		request.SetPathValue("metric_type", v.mType)
		request.SetPathValue("metric_name", v.mName)
		require.NoError(t, err)

		response := httptest.NewRecorder()
		mh.GetMetricByValue(response, request)
		result := response.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusNotFound, response.Code)
		require.Equal(t, "No such metric type", response.Body.String())
	}
}

// ------- Хендлер: GET /
// r.Get("/", mware.WithLogging(mware.GzipMiddleware(mh.GetAllMetrics)))
func TestGetAllMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	expectedBody := `
	<html>      

        <body>
            <h1>Gauge metrics</h1>
        	<table>
    
                <tr>
                    <td>g1</td>
                    <td>1.1</td>
                </tr>
    
                <tr>
                    <td>g2</td>
                    <td>2.2</td>
                </tr>
    
            </table>
    
            <h1>Counter metrics</h1>
            <table>
    
                <tr>
                    <td>c1</td>
                    <td>1</td>
                </tr>
    
                <tr>
                    <td>c2</td>
                    <td>10</td>
                </tr>
    
            </table>
        </body>
                                
    </html>`

	m.EXPECT().GetAllGaugeMetrics().Return(map[string]float64{"g1": 1.1, "g2": 2.2}, nil)
	m.EXPECT().GetAllCounterMetrics().Return(map[string]int64{"c1": 1, "c2": 10}, nil)
	response := httptest.NewRecorder()

	err := service.WriteMetricsReport(m, response)
	require.NoError(t, err)

	contentType := response.Header().Get("Content-type")
	require.Equal(t, http.StatusOK, response.Code)
	require.True(t, strings.Contains(contentType, "text/html"))
	require.Equal(t, remoweWs(expectedBody), remoweWs(response.Body.String()))
}

func TestGetAllMetricsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	reqURL := "/"

	expectedBody := `
	<html>      

        <body>
            <h1>Gauge metrics</h1>
        	<table>
    
                <tr>
                    <td>g1</td>
                    <td>1.1</td>
                </tr>
    
                <tr>
                    <td>g2</td>
                    <td>2.2</td>
                </tr>
    
            </table>
    
            <h1>Counter metrics</h1>
            <table>
    
                <tr>
                    <td>c1</td>
                    <td>1</td>
                </tr>
    
                <tr>
                    <td>c2</td>
                    <td>10</td>
                </tr>
    
            </table>
        </body>
                                
    </html>`

	m.EXPECT().GetAllGaugeMetrics().Return(map[string]float64{"g1": 1.1, "g2": 2.2}, nil)
	m.EXPECT().GetAllCounterMetrics().Return(map[string]int64{"c1": 1, "c2": 10}, nil)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.GetAllMetrics(response, request)

	contentType := response.Header().Get("Content-type")
	require.Equal(t, http.StatusOK, response.Code)
	require.True(t, strings.Contains(contentType, "text/html"))
	require.Equal(t, remoweWs(expectedBody), remoweWs(response.Body.String()))
}

func remoweWs(s string) string {
	rpl := strings.NewReplacer("\n", "", " ", "", "\t", "")
	return rpl.Replace(s)
}

// r.Get("/ping", mware.WithLogging(mware.GzipMiddleware(mh.CheckConnectionToDB)))
func TestCheckConnectionToDBFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mDB := mocks.NewMockDBConnector(ctrl)

	mh := &MetricHandlers{
		DBCon: mDB,
	}

	reqURL := "/ping"

	mDB.EXPECT().Connect().Return(nil, fmt.Errorf("Connection error"))
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	require.NoError(t, err)

	response := httptest.NewRecorder()
	mh.CheckConnectionToDB(response, request)

	require.Equal(t, http.StatusInternalServerError, response.Code)
}

// r.Get("/ping", mware.WithLogging(mware.GzipMiddleware(mh.CheckConnectionToDB)))
func TestCheckConnectionToDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mDB := mocks.NewMockDBConnector(ctrl)

	mh := &MetricHandlers{
		DBCon: mDB,
	}

	reqURL := "/ping"

	mDB.EXPECT().Connect().Return(nil, nil)

	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.CheckConnectionToDB(response, request)
	require.Equal(t, http.StatusOK, response.Code)
}

// r.Post("/update/{metric_type}/{metric_name}/{metric_value}", mware.WithLogging(mh.CreateMetric))
func TestCreateMetricCounterTypeFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	mType := "counter"
	mName := "c1"
	mValue := "100"

	value, err := strconv.ParseInt(mValue, 10, 64)
	require.NoError(t, err)

	reqURL := fmt.Sprintf("/update/%s/%s/%s", mType, mName, mValue)

	m.EXPECT().AddCounterMetric(mName, value).Return(fmt.Errorf("Error in adding counter metric"))

	request, err := http.NewRequest(http.MethodPost, reqURL, nil)
	request.SetPathValue("metric_type", mType)
	request.SetPathValue("metric_name", mName)
	request.SetPathValue("metric_value", mValue)
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateMetricCounterTypeWrongValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	mType := "counter"
	mName := "c1"
	mValue := "abc"

	reqURL := fmt.Sprintf("/update/%s/%s/%s", mType, mName, mValue)

	request, err := http.NewRequest(http.MethodPost, reqURL, nil)
	request.SetPathValue("metric_type", mType)
	request.SetPathValue("metric_name", mName)
	request.SetPathValue("metric_value", mValue)
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestCreateMetricCounterType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	mType := "counter"
	mName := "c1"
	mValue := "100"

	value, err := strconv.ParseInt(mValue, 10, 64)
	require.NoError(t, err)

	reqURL := fmt.Sprintf("/update/%s/%s/%s", mType, mName, mValue)

	m.EXPECT().AddCounterMetric(mName, value).Return(nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, nil)
	request.SetPathValue("metric_type", mType)
	request.SetPathValue("metric_name", mName)
	request.SetPathValue("metric_value", mValue)
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "Created", response.Body.String())
}

func TestCreateMetricGaugeTypeFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	mType := "gauge"
	mName := "g1"
	mValue := "100.111"
	value, err := strconv.ParseFloat(mValue, 64)
	require.NoError(t, err)

	reqURL := fmt.Sprintf("/update/%s/%s/%s", mType, mName, mValue)

	m.EXPECT().AddGaugeMetric(mName, value).Return(fmt.Errorf("Error in adding gauge metric"))

	request, err := http.NewRequest(http.MethodPost, reqURL, nil)
	request.SetPathValue("metric_type", mType)
	request.SetPathValue("metric_name", mName)
	request.SetPathValue("metric_value", mValue)
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateMetricGaugeTypeWrongValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	mType := "gauge"
	mName := "g1"
	mValue := "abc"

	reqURL := fmt.Sprintf("/update/%s/%s/%s", mType, mName, mValue)

	request, err := http.NewRequest(http.MethodPost, reqURL, nil)
	request.SetPathValue("metric_type", mType)
	request.SetPathValue("metric_name", mName)
	request.SetPathValue("metric_value", mValue)
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestCreateMetricGaugeType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	mType := "gauge"
	mName := "g1"
	mValue := "100.111"

	value, err := strconv.ParseFloat(mValue, 64)
	require.NoError(t, err)

	reqURL := fmt.Sprintf("/update/%s/%s/%s", mType, mName, mValue)

	m.EXPECT().AddGaugeMetric(mName, value).Return(nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, nil)
	request.SetPathValue("metric_type", mType)
	request.SetPathValue("metric_name", mName)
	request.SetPathValue("metric_value", mValue)
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "Created", response.Body.String())
}

// r.Post("/update/", mware.WithLogging(mware.GzipMiddleware(mh.MetricUpdate)))
func TestMetricUpdateCounterMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	reqURL := "/update/"

	mName := "C1"
	mValue := "123"
	value, err := strconv.ParseInt(mValue, 10, 64)
	require.NoError(t, err)
	reqBody := `{"id":"C1", "type":"counter", "delta":123}`

	m.EXPECT().AddCounterMetric(mName, value).Return(nil)
	m.EXPECT().GetCounterMetric(mName).Return(value, true, nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	// assert.Equal(t, http.MethodPost, request.Method) //
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.MetricUpdate(response, request)
	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, reqBody, string(resBody))
}

func TestMetricUpdateIncorrectMetricFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	reqURL := "/update/"

	reqBody := `{"id":"C1", "type":"wrongType", "delta":123}`

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.MetricUpdate(response, request)
	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "No such metric", string(resBody))
}

func TestMetricUpdateCounterMetricInvalidJSONFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	reqURL := "/update/"

	reqBody := `{"id":"C1", "type":"counter", "delta":111.222}`

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.MetricUpdate(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestMetricUpdateGaugeMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	reqURL := "/update/"

	mName := "G1"
	mValue := "111.333"
	value, err := strconv.ParseFloat(mValue, 64)
	require.NoError(t, err)
	reqBody := `{"id":"G1", "type":"gauge", "value":111.333}`

	m.EXPECT().AddGaugeMetric(mName, value).Return(nil)
	m.EXPECT().GetGaugeMetric(mName).Return(value, true, nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.MetricUpdate(response, request)
	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, reqBody, string(resBody))
}

// r.Post("/value/", mware.WithLogging(mware.GzipMiddleware(mh.MetricValue)))
func TestMetricValueCounterMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	reqURL := "/value/"

	mName := "C1"
	var value int64 = 123

	reqBody := `{"id":"C1", "type":"counter"}`
	expectedRespBody := `{"id":"C1", "type":"counter", "delta":123}`
	m.EXPECT().GetCounterMetric(mName).Return(value, true, nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.MetricValue(response, request)
	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, expectedRespBody, string(resBody))
}

func TestMetricValueGaugeMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
	}

	reqURL := "/value/"

	mName := "G1"
	value := 111.333

	reqBody := `{"id":"G1", "type":"gauge"}`
	expectedRespBody := `{"id":"G1", "type":"gauge", "value":111.333}`
	m.EXPECT().GetGaugeMetric(mName).Return(value, true, nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.MetricValue(response, request)
	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, expectedRespBody, string(resBody))
}

// r.Post("/updates/", mware.WithLogging(mware.GzipMiddleware(mh.MetricsUpdate)))
func TestMetricsUpdateCounterMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
		Config: &config.Config{
			Key: "secret",
		},
	}

	reqURL := "/update/"

	reqBody := `[{"id":"c1", "type":"counter", "delta":5}, {"id":"c2", "type":"counter", "delta":8}]`

	hash := security.CreateSign(reqBody, mh.Config.Key)
	d1 := int64(5)
	d2 := int64(8)

	metrics := []memstorage.Metric{
		{
			ID:    "c1",
			MType: "counter",
			Delta: &d1,
		},
		{
			ID:    "c2",
			MType: "counter",
			Delta: &d2,
		},
	}

	for _, metric := range metrics {
		mm := metric
		m.EXPECT().AddCounterMetric(mm.ID, *mm.Delta).Return(nil)
		m.EXPECT().GetCounterMetric(mm.ID).Return(*mm.Delta, true, nil)
	}
	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	request.Header.Set("HashSHA256", hash)

	response := httptest.NewRecorder()

	mh.MetricsUpdate(response, request)
	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"result":"ok"}`, string(resBody))

}

func TestMetricsUpdateGaugeMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
		Config: &config.Config{
			Key: "secret",
		},
	}

	reqURL := "/update/"

	reqBody := `[{"id":"c1", "type":"gauge", "value":1.1}, {"id":"c2", "type":"gauge", "value":2.222}]`

	hash := security.CreateSign(reqBody, mh.Config.Key)
	d1 := 1.1
	d2 := 2.222

	metrics := []memstorage.Metric{
		{
			ID:    "c1",
			MType: "gauge",
			Value: &d1,
		},
		{
			ID:    "c2",
			MType: "gauge",
			Value: &d2,
		},
	}

	for _, metric := range metrics {
		mm := metric
		m.EXPECT().AddGaugeMetric(mm.ID, *mm.Value).Return(nil)
		m.EXPECT().GetGaugeMetric(mm.ID).Return(*mm.Value, true, nil)
	}
	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	request.Header.Set("HashSHA256", hash)

	response := httptest.NewRecorder()

	mh.MetricsUpdate(response, request)
	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"result":"ok"}`, string(resBody))
}

func TestMetricsUpdateCounterMetricIncorrectSignatureFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mh := &MetricHandlers{
		Config: &config.Config{
			Key: "secret",
		},
	}

	reqURL := "/update/"

	reqBody := `[{"id":"c1", "type":"counter", "delta":5}, {"id":"c2", "type":"counter", "delta":8}]`

	incorrectKey := "wrongsecret"
	hash := security.CreateSign(reqBody, incorrectKey)

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	request.Header.Set("HashSHA256", hash)

	response := httptest.NewRecorder()

	mh.MetricsUpdate(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
}
