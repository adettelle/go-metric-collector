package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// структура для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}
	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// WithLogging добавляет дополнительный код для регистрации сведений о запросе
// и возвращает новый http.Handler.
func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logger, err := zap.NewDevelopment()
		if err != nil {
			// вызываем панику, если ошибка
			panic("cannot initialize zap")
		}
		defer logger.Sync()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		logger.Info("Request data:", zap.String("uri", r.RequestURI),
			zap.String("method", r.Method), zap.Int("status", responseData.status),
			zap.Duration("duration", duration), zap.Int("size", responseData.size))

	}
	// возвращаем функционально расширенный хендлер
	return http.HandlerFunc(logFn)
}
