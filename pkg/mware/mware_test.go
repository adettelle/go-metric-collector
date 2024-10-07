package mware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func compress(data []byte, t *testing.T) *bytes.Buffer {

	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	_, err := zb.Write([]byte(data)) // в buf будет заархивированные данные data

	require.NoError(t, err) // то и есть обработка ошибки выше!!!!
	err = zb.Close()
	require.NoError(t, err)

	return buf
}

func decompress(data []byte, t *testing.T) string { // io.ReadCloser, t *testing.T; *bytes.Buffer

	reader := bytes.NewReader(data)
	gzreader, err := gzip.NewReader(reader) // в gzreader лежит источник для разархивирования
	// при начале чтения начнется разархивирование

	require.NoError(t, err) // то и есть обработка ошибки выше!!!!

	output, err := io.ReadAll(gzreader)
	if err != nil {
		fmt.Println(err)
	}

	result := string(output)
	return result
}

func TestGzipMiddlewareNoCompression(t *testing.T) {
	handler := http.HandlerFunc(GzipMiddleware(EchoHandler))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 0.999888777
	}`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 0.999888777
	}`

	buf := bytes.NewBufferString(requestBody)
	r := httptest.NewRequest("POST", srv.URL, buf)
	r.RequestURI = ""

	resp, err := http.DefaultClient.Do(r)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.JSONEq(t, successBody, string(b))
}

// запрос перед отправкой надо заgzip'овать
// middleware получает заархивированные данные, разархивирует их (сам, мы здесь ничего не делаем)
// потом он передает уже обычный json в хэндлер, хэндлер отвечает тем же самым json'ом
// далее возвращаемый json перехватывает middleware и ничего с ним не делает,
// передает клиенту в неизменном виде
// потому что мы не ставили хэддер Accept-Encoding
func TestGzipMiddlewareCompressRequest(t *testing.T) {
	handler := http.HandlerFunc(GzipMiddleware(EchoHandler))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 1.111111
	}`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 1.111111
	}`

	buf := compress([]byte(requestBody), t)

	r := httptest.NewRequest("POST", srv.URL, buf)
	r.RequestURI = ""
	r.Header.Set("Content-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r) // это то, что вернулось от сервера
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.JSONEq(t, successBody, string(b))
}

// Incorrect Content-Type test
func TestGzipMiddlewareCompressBadRequest(t *testing.T) {
	handler := http.HandlerFunc(GzipMiddleware(EchoHandler))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 1.111111
	}`

	buf := compress([]byte(requestBody), t)

	r := httptest.NewRequest("POST", srv.URL, buf)
	r.RequestURI = ""
	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set("Content-Type", "application/xml")

	resp, err := http.DefaultClient.Do(r) // это то, что вернулось от сервера
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	defer resp.Body.Close()
}

// запрос перед отправкой не надо заgzip'овать
// middleware получает незаархивированные данные
// потом он передает обычный json в хэндлер, хэндлер отвечает тем же самым json'ом
// далее возвращаемый json перехватывает middleware и архивирует его и
// передает клиенту в заархивированном виде
// потому что стоит хэддер Accept-Encoding
func TestGzipMiddlewareCompressResponse(t *testing.T) {
	handler := http.HandlerFunc(GzipMiddleware(EchoHandler))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 0.1
	}`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 0.1
	}`

	r := httptest.NewRequest("POST", srv.URL, bytes.NewBuffer([]byte(requestBody)))
	r.RequestURI = ""
	r.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r) // это то, что вернулось от сервера
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	res := decompress(b, t)

	require.JSONEq(t, successBody, res)
}

// запрос перед отправкой надо заgzip'овать
// middleware получает заархивированные данные, разархивирует их (сам, мы здесь ничего не делаем)
// потом он передает уже обычный json в хэндлер, хэндлер отвечает тем же самым json'ом
// далее возвращаемый json перехватывает middleware и архивирует его и
// передает клиенту в заархивированном виде
// потому что стоит хэддер Accept-Encoding
func TestGzipMiddlewareCompressAll(t *testing.T) {
	handler := http.HandlerFunc(GzipMiddleware(EchoHandler))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 5.1
	}`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 5.1
	}`

	buf := compress([]byte(requestBody), t)

	r := httptest.NewRequest("POST", srv.URL, buf)
	r.RequestURI = ""
	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r) // это то, что вернулось от сервера, оно заархивированно
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	res := decompress(b, t)

	require.JSONEq(t, successBody, res)
}

// запрос перед отправкой не надо заgzip'овать
// но говорю headder'ом, что они заархивированные Content-Encoding
// тогда middleware не сможет их разархивировать и должен дать ошибку, 500 статус, мы его ожидаем
func TestGzipMiddlewareIncorrectCompressRequest(t *testing.T) {
	handler := http.HandlerFunc(GzipMiddleware(EchoHandler))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{
		"id":"Alloc", 
		"type": "gauge", 
		"value": 1.1
	}`

	r := httptest.NewRequest("POST", srv.URL, bytes.NewBuffer([]byte(requestBody)))
	r.RequestURI = ""
	r.Header.Set("Content-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r) // это то, что вернулось от сервера
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	defer resp.Body.Close()
}

// достает из запроса body и записывает его в ответ
func EchoHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	w.Write(b)
	w.WriteHeader(http.StatusOK)
}
