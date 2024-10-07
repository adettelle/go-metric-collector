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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mertic[T any] struct {
	Type   string
	Name   string
	Value  T
	Exists bool
}

func CounterMetric(name string, value int64) mertic[int64] {
	mc := mertic[int64]{
		Type:  "counter",
		Name:  name,
		Value: value,
	}
	return mc
}

func GaugeMetric(name string, value float64) mertic[float64] {
	mg := mertic[float64]{
		Type:  "gauge",
		Name:  name,
		Value: value,
	}
	return mg
}

func CreateMetricHandlers(t *testing.T) (*MetricHandlers, func()) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	// создаём объект-заглушку
	m := mocks.NewMockStorager(ctrl)

	mh := &MetricHandlers{
		Storager: m,
		Config:   nil,
	}
	return mh, ctrl.Finish
}

func CreateRequestWithPathValues(t *testing.T, method string, url string, body io.Reader, mType string, mName string) *http.Request {
	request, err := http.NewRequest(method, url, body)
	request.SetPathValue("metric_type", mType)
	request.SetPathValue("metric_name", mName)
	require.NoError(t, err)
	return request
}

// ------- Хендлер: GET /value/{metric_type}/{metric_name}
// test case with no such counter metric
func TestGetMetricByValueWithNoCounterMetric(t *testing.T) {
	mh, cleanUp := CreateMetricHandlers(t)
	defer cleanUp()

	mc := CounterMetric("C1", 0)
	reqURL := fmt.Sprintf("/value/%s/%s", mc.Type, mc.Name)

	m := mh.Storager.(*mocks.MockStorager)
	m.EXPECT().GetCounterMetric(mc.Name).Return(mc.Value, mc.Exists, nil)

	request := CreateRequestWithPathValues(t, http.MethodGet, reqURL, nil, mc.Type, mc.Name)
	response := httptest.NewRecorder()
	mh.GetMetricByValue(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
}

// test case with error in getting counter metric
func TestGetMetricByValueWithErrorInGettingCounterMetric(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	defer cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mc := CounterMetric("C1", 0)
	metricExists := false

	reqURL := fmt.Sprintf("/value/%s/%s", mc.Type, mc.Name)

	m.EXPECT().GetCounterMetric(mc.Name).Return(mc.Value, metricExists, fmt.Errorf("Error in getting counter metric"))

	request := CreateRequestWithPathValues(t, http.MethodGet, reqURL, nil, mc.Type, mc.Name)
	response := httptest.NewRecorder()
	mh.GetMetricByValue(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

// test case with counter metric
func TestGetMetricCounterByValue(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	defer cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mc := CounterMetric("C1", 10)
	metricExists := true

	reqURL := fmt.Sprintf("/value/%s/%s", mc.Type, mc.Name)

	// пишем, что хотим получить от заглушки
	m.EXPECT().GetCounterMetric(mc.Name).Return(mc.Value, metricExists, nil)

	// тестируем хэндлер r.Get("/value/{metric_type}/{metric_name}", mware.WithLogging(mh.GetMetricByValue))
	// 1. Создаем запрос для обработчика
	request := CreateRequestWithPathValues(t, http.MethodGet, reqURL, nil, mc.Type, mc.Name)
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

// test case with no such gauge metric
func TestGetMetricByValueWithNoGaugeMetric(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	defer cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg := GaugeMetric("G1", 0)

	metricExists := false

	reqURL := fmt.Sprintf("/value/%s/%s", mg.Type, mg.Name)

	m.EXPECT().GetGaugeMetric(mg.Name).Return(mg.Value, metricExists, nil)

	request := CreateRequestWithPathValues(t, http.MethodGet, reqURL, nil, mg.Type, mg.Name)
	response := httptest.NewRecorder()
	mh.GetMetricByValue(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
}

// test case with error in getting gauge metric
func TestGetMetricByValueWithErrorInGettingGaugeMetric(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	defer cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg := GaugeMetric("G1", 0)
	metricExists := false

	reqURL := fmt.Sprintf("/value/%s/%s", mg.Type, mg.Name)

	m.EXPECT().GetGaugeMetric(mg.Name).Return(mg.Value, metricExists, fmt.Errorf("Error in getting gauge metric"))

	request := CreateRequestWithPathValues(t, http.MethodGet, reqURL, nil, mg.Type, mg.Name)
	response := httptest.NewRecorder()
	mh.GetMetricByValue(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

// test case with gauge metric
func TestGetMetricGaugeByValue(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	defer cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg1 := GaugeMetric("G1", 123.0)
	mg2 := GaugeMetric("G2", 150984.573)

	testTable := []mertic[float64]{mg1, mg2}
	mExists := true // !!!!!!!!!!!!!!!!

	for _, mg := range testTable {
		reqURL := fmt.Sprintf("/value/%s/%s", mg.Type, mg.Name)

		m.EXPECT().GetGaugeMetric(mg.Name).Return(mg.Value, mExists, nil) // !!!!!!!!!!!!

		request := CreateRequestWithPathValues(t, http.MethodGet, reqURL, nil, mg.Type, mg.Name)
		response := httptest.NewRecorder()
		mh.GetMetricByValue(response, request)
		result := response.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusOK, response.Code)
		n, err := strconv.ParseFloat(response.Body.String(), 64)
		require.NoError(t, err)
		require.Equal(t, mg.Value, n)
	}
}

// проверим на ошибочный запрос
func TestGetMetricByWrongValue(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	defer cleanup()

	mType := "somemetric"
	mName := "a6"

	reqURL := fmt.Sprintf("/value/%s/%s", mType, mName)

	// тестируем хэндлер r.Get("/value/{metric_type}/{metric_name}", mware.WithLogging(mh.GetMetricByValue))
	request := CreateRequestWithPathValues(t, http.MethodGet, reqURL, nil, mType, mName)
	response := httptest.NewRecorder()
	mh.GetMetricByValue(response, request)
	result := response.Result()
	defer result.Body.Close()

	require.Equal(t, http.StatusNotFound, response.Code)
	require.Equal(t, "No such metric type", response.Body.String())
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
func TestCreateMetricWrongMethod(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	mc := CounterMetric("c1", 100)
	reqURL := fmt.Sprintf("/update/%s/%s/%d", mc.Type, mc.Name, mc.Value)

	request := CreateRequestWithPathValues(t, http.MethodGet, reqURL, nil, mc.Type, mc.Name)
	response := httptest.NewRecorder()
	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusMethodNotAllowed, response.Code)
}

func TestCreateMetricCounterTypeFail(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mc := CounterMetric("c1", 100)
	reqURL := fmt.Sprintf("/update/%s/%s/%d", mc.Type, mc.Name, mc.Value)

	m.EXPECT().AddCounterMetric(mc.Name, mc.Value).Return(fmt.Errorf("Error in adding counter metric"))

	request := CreateRequestWithPathValues(t, http.MethodPost, reqURL, nil, mc.Type, mc.Name)
	request.SetPathValue("metric_value", strconv.FormatInt(mc.Value, 10))
	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateMetricCounterTypeWrongValue(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	defer cleanup()

	mType := "counter"
	mName := "c1"
	mValue := "abc"

	reqURL := fmt.Sprintf("/update/%s/%s/%s", mType, mName, mValue)

	request := CreateRequestWithPathValues(t, http.MethodPost, reqURL, nil, mType, mName)
	request.SetPathValue("metric_value", mValue)
	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestCreateMetricCounterType(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mc := CounterMetric("c1", 100)
	reqURL := fmt.Sprintf("/update/%s/%s/%d", mc.Type, mc.Name, mc.Value)

	m.EXPECT().AddCounterMetric(mc.Name, mc.Value).Return(nil)

	request := CreateRequestWithPathValues(t, http.MethodPost, reqURL, nil, mc.Type, mc.Name)
	request.SetPathValue("metric_value", strconv.FormatInt(mc.Value, 10))
	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "Created", response.Body.String())
}

func TestCreateMetricGaugeTypeFail(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg := GaugeMetric("g1", 100.111)
	reqURL := fmt.Sprintf("/update/%s/%s/%f", mg.Type, mg.Name, mg.Value)

	m.EXPECT().AddGaugeMetric(mg.Name, mg.Value).Return(fmt.Errorf("Error in adding gauge metric"))

	request := CreateRequestWithPathValues(t, http.MethodPost, reqURL, nil, mg.Type, mg.Name)
	request.SetPathValue("metric_value", fmt.Sprintf("%f", mg.Value))
	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateMetricGaugeTypeWrongValue(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	mType := "gauge"
	mName := "g1"
	mValue := "abc"

	reqURL := fmt.Sprintf("/update/%s/%s/%s", mType, mName, mValue)

	request := CreateRequestWithPathValues(t, http.MethodPost, reqURL, nil, mType, mName)
	request.SetPathValue("metric_value", mValue)
	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestCreateMetricGaugeType(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg := GaugeMetric("g1", 100.111)
	reqURL := fmt.Sprintf("/update/%s/%s/%f", mg.Type, mg.Name, mg.Value)

	m.EXPECT().AddGaugeMetric(mg.Name, mg.Value).Return(nil)

	request := CreateRequestWithPathValues(t, http.MethodPost, reqURL, nil, mg.Type, mg.Name)
	request.SetPathValue("metric_value", fmt.Sprintf("%f", mg.Value))
	response := httptest.NewRecorder()

	mh.CreateMetric(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "Created", response.Body.String())
}

// r.Post("/update/", mware.WithLogging(mware.GzipMiddleware(mh.MetricUpdate)))
func TestMetricUpdateCounterMetric(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mc := CounterMetric("c1", 123)
	reqURL := "/update/"
	reqBody := `{"id":"c1", "type":"counter", "delta":123}`

	m.EXPECT().AddCounterMetric(mc.Name, mc.Value).Return(nil)
	m.EXPECT().GetCounterMetric(mc.Name).Return(mc.Value, true, nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	response := httptest.NewRecorder()

	mh.MetricUpdate(response, request)
	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, reqBody, string(resBody))
}

func TestMetricUpdateIncorrectMetricFail(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	reqURL := "/update/"
	reqBody := `{"id":"c1", "type":"wrongType", "delta":123}`
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
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	reqURL := "/update/"
	reqBody := `{"id":"C1", "type":"counter", "delta":111.222}`
	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)

	response := httptest.NewRecorder()

	mh.MetricUpdate(response, request)
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestMetricUpdateGaugeMetric(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg := GaugeMetric("G1", 111.333)
	reqURL := "/update/"
	reqBody := `{"id":"G1", "type":"gauge", "value":111.333}`

	m.EXPECT().AddGaugeMetric(mg.Name, mg.Value).Return(nil)
	m.EXPECT().GetGaugeMetric(mg.Name).Return(mg.Value, true, nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	response := httptest.NewRecorder()

	mh.MetricUpdate(response, request)
	resBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, reqBody, string(resBody))
}

func TestMetricUpdateAddCounterMetricFail(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mc := CounterMetric("C1", 123)
	reqURL := "/update/"
	reqBody := `{"id":"C1", "type":"counter", "delta":123}`

	m.EXPECT().AddCounterMetric(mc.Name, mc.Value).Return(fmt.Errorf("Error in adding counter metric"))

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	response := httptest.NewRecorder()
	mh.MetricUpdate(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestMetricUpdateGetCounterMetricFail(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mc := CounterMetric("C1", 123)
	reqURL := "/update/"
	reqBody := `{"id":"C1", "type":"counter", "delta":123}`

	m.EXPECT().AddCounterMetric(mc.Name, mc.Value).Return(nil)
	m.EXPECT().GetCounterMetric(mc.Name).Return(mc.Value, true, fmt.Errorf("Error in getting counter metric"))

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	response := httptest.NewRecorder()
	mh.MetricUpdate(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestMetricUpdateGetCounterMetricNotOK(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mc := CounterMetric("C1", 123)

	reqURL := "/update/"
	reqBody := `{"id":"C1", "type":"counter", "delta":123}`

	m.EXPECT().AddCounterMetric(mc.Name, mc.Value).Return(nil)
	m.EXPECT().GetCounterMetric(mc.Name).Return(mc.Value, false, nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	response := httptest.NewRecorder()
	mh.MetricUpdate(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestMetricUpdateAddGaugeMetricFail(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg := GaugeMetric("G1", 123.111)

	reqURL := "/update/"
	reqBody := `{"id":"G1", "type":"gauge", "value":123.111}`

	m.EXPECT().AddGaugeMetric(mg.Name, mg.Value).Return(fmt.Errorf("Error in adding gauge metric"))

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	response := httptest.NewRecorder()
	mh.MetricUpdate(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestMetricUpdateGetGaugeMetricFail(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg := GaugeMetric("G1", 123.111)

	reqURL := "/update/"
	reqBody := `{"id":"G1", "type":"gauge", "value":123.111}`

	m.EXPECT().AddGaugeMetric(mg.Name, mg.Value).Return(nil)
	m.EXPECT().GetGaugeMetric(mg.Name).Return(mg.Value, true, fmt.Errorf("Error in getting gauge metric"))

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	response := httptest.NewRecorder()
	mh.MetricUpdate(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestMetricUpdateGetGaugeMetricNotOK(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg := GaugeMetric("G1", 123.111)
	reqURL := "/update/"
	reqBody := `{"id":"G1", "type":"gauge", "value":123.111}`

	m.EXPECT().AddGaugeMetric(mg.Name, mg.Value).Return(nil)
	m.EXPECT().GetGaugeMetric(mg.Name).Return(mg.Value, false, nil)

	request, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	require.NoError(t, err)
	response := httptest.NewRecorder()
	mh.MetricUpdate(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
}

// r.Post("/value/", mware.WithLogging(mware.GzipMiddleware(mh.MetricValue)))
func TestMetricValueCounterMetric(t *testing.T) {
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mc := CounterMetric("C1", 123)

	reqURL := "/value/"
	reqBody := `{"id":"C1", "type":"counter"}`
	expectedRespBody := `{"id":"C1", "type":"counter", "delta":123}`
	m.EXPECT().GetCounterMetric(mc.Name).Return(mc.Value, true, nil)

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
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()

	m := mh.Storager.(*mocks.MockStorager)
	mg := GaugeMetric("G1", 111.333)

	reqURL := "/value/"
	reqBody := `{"id":"G1", "type":"gauge"}`
	expectedRespBody := `{"id":"G1", "type":"gauge", "value":111.333}`
	m.EXPECT().GetGaugeMetric(mg.Name).Return(mg.Value, true, nil)

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
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()
	mh.Config = &config.Config{Key: "secret"}

	m := mh.Storager.(*mocks.MockStorager)

	reqURL := "/update/"
	reqBody := `[{"id":"c1", "type":"counter", "delta":5}, {"id":"c2", "type":"counter", "delta":8}]`

	hash := security.CreateSign(reqBody, mh.Config.Key)
	mc1 := CounterMetric("c1", 5)
	mc2 := CounterMetric("c2", 8)

	metrics := []mertic[int64]{mc1, mc2}
	for _, metric := range metrics {
		mm := metric
		m.EXPECT().AddCounterMetric(mm.Name, mm.Value).Return(nil)
		m.EXPECT().GetCounterMetric(mm.Name).Return(mm.Value, true, nil)
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
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()
	mh.Config = &config.Config{Key: "secret"}

	m := mh.Storager.(*mocks.MockStorager)

	reqURL := "/update/"
	reqBody := `[{"id":"g1", "type":"gauge", "value":1.1}, {"id":"g2", "type":"gauge", "value":2.222}]`

	hash := security.CreateSign(reqBody, mh.Config.Key)
	mg1 := GaugeMetric("g1", 1.1)
	mg2 := GaugeMetric("g2", 2.222)

	metrics := []mertic[float64]{mg1, mg2}
	for _, metric := range metrics {
		mm := metric
		m.EXPECT().AddGaugeMetric(mm.Name, mm.Value).Return(nil)
		m.EXPECT().GetGaugeMetric(mm.Name).Return(mm.Value, true, nil)
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
	mh, cleanup := CreateMetricHandlers(t)
	cleanup()
	mh.Config = &config.Config{Key: "secret"}

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
