// пакет api с описанием структурных типов с конструкторами:
// Handler, Router (Server?) всё что связано с сервером
package api

import (
	"github.com/adettelle/go-metric-collector/internal/handlers"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
	"github.com/go-chi/chi/v5"
)

func NewMetricRouter(ms *memstorage.MemStorage, mAPI *handlers.MetricHandlers) *chi.Mux {

	r := chi.NewRouter()

	// POST /update/counter/someMetric/123
	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", mAPI.CreateMetric)
	r.Get("/value/{metric_type}/{metric_name}", mAPI.GetMetricByValue)
	r.Get("/", mAPI.GetAllMetrics)

	return r
}
