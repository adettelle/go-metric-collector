// пакет api с описанием структурных типов с конструкторами:
// Handler, Router (Server?) всё что связано с сервером
package api

import (
	"github.com/adettelle/go-metric-collector/internal/handlers"
	"github.com/adettelle/go-metric-collector/internal/logger"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
	"github.com/go-chi/chi/v5"
)

func NewMetricRouter(ms *memstorage.MemStorage, mh *handlers.MetricHandlers) *chi.Mux {

	r := chi.NewRouter()

	// POST http://localhost:8080/update/counter/someMetric/123
	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", logger.WithLogging(mh.CreateMetric))
	r.Get("/value/{metric_type}/{metric_name}", logger.WithLogging(mh.GetMetricByValue))
	r.Get("/", logger.WithLogging(mh.GetAllMetrics))

	r.Post("/update/", logger.WithLogging(mh.JSONHandlerUpdate))
	r.Post("/value/", logger.WithLogging(mh.JSONHandlerValue))

	return r
}
