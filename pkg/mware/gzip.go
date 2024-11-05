// Package mware provides middleware functionality for HTTP servers.
// This package contains utilities such as middleware to handle compression,
// specifically gzip, for incoming and outgoing HTTP requests.
// This package also offers functionality such as logging HTTP request/response data,
// including status codes, request durations, and response sizes.
// It also includes a custom implementation of the http.ResponseWriter
// to capture detailed information about the HTTP response.
package mware

import (
	"net/http"
	"strings"

	pkg "github.com/adettelle/go-metric-collector/pkg/compressor"
)

// GzipMiddleware wraps an http.HandlerFunc to add gzip compression and decompression functionality.
// It compresses responses and decompresses requests
// based on the client's `Accept-Encoding` and `Content-Encoding` headers.
func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		contentType := r.Header.Values("Content-Type")
		for _, ct := range contentType {
			if !strings.Contains(ct, "application/json") &&
				!strings.Contains(ct, "text/html") {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Values("Accept-Encoding")
		for _, ae := range acceptEncoding {
			supportsGzip := strings.Contains(ae, "gzip")
			if supportsGzip {
				w.Header().Set("Content-Encoding", "gzip")
				// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
				cw := pkg.NewCompressWriter(w)
				// меняем оригинальный http.ResponseWriter на новый
				ow = cw
				// не забываем отправить клиенту все сжатые данные после завершения middleware
				defer cw.Close()
			}
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Values("Content-Encoding")
		for _, ce := range contentEncoding {
			sendsGzip := strings.Contains(ce, "gzip")
			if sendsGzip {
				// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
				cr, err := pkg.NewCompressReader(r.Body)

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// меняем тело запроса на новое
				r.Body = cr
				defer cr.Close()
			}
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
}
