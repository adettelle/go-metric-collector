// пакет api с описанием структурных типов с конструкторами:
// Handler, Router (Server?) всё что связано с сервером
package api

import (
	"github.com/adettelle/go-metric-collector/pkg/mware"
	"github.com/go-chi/chi/v5"
)

func NewMetricRouter(ms Storager, mh *MetricHandlers) *chi.Mux { // ms *memstorage.MemStorage

	r := chi.NewRouter()

	// POST http://localhost:8080/update/counter/someMetric/123
	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", mware.WithLogging(mh.CreateMetric))
	r.Get("/value/{metric_type}/{metric_name}", mware.WithLogging(mh.GetMetricByValue))

	r.Get("/", mware.WithLogging(mware.GzipMiddleware(mh.GetAllMetrics)))

	// метод получает метрику на вход для обновления и для добавления
	// GzipMiddleware смотрит на HTTP-заголовка Content-Encoding
	// и разархивирует body (если gzip) либо оставляет, как есть
	// принимает в теле запроса метрику в формате json
	r.Post("/update/", mware.WithLogging(mware.GzipMiddleware(mh.MetricUpdate)))

	// метод отдает значение метрики
	// GzipMiddleware смотрит на заголовок Accept-Encoding
	// и если он gzip, то перед записью ответа сжимает его
	r.Post("/value/", mware.WithLogging(mware.GzipMiddleware(mh.MetricValue)))
	r.Get("/ping", mware.WithLogging(mware.GzipMiddleware(mh.CheckConnectionToDB)))

	// принимает в теле запроса множество метрик в формате: []Metrics (списка метрик) в виде json
	r.Post("/updates/", mware.WithLogging(mware.GzipMiddleware(mh.MetricsUpdate)))

	return r
}
