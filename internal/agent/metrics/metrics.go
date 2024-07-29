// слой хранения: отвечает за хранение метрик (подразумевается два дериватива: положить и достать)
// аналог банковской ячейки (положить, достать)
package metrics

import (
	"sync"
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type AllMetrics struct {
	AllMetrics []Metric `json:"metrics"`
}

// MetricAccumulator is used for storaging metrics
type MetricAccumulator struct {
	sync.RWMutex
	gauge   map[string]float64 // имя метрики: ее значение
	counter map[string]int64
}

// оставляем
// Reset() обнуляет карты Gauge и Counter в структуре MemStorage
// метод применяется после отправки всех метрик
func (ma *MetricAccumulator) Reset() {
	ma.Lock()
	defer ma.Unlock()

	for k := range ma.gauge {
		delete(ma.gauge, k)
	}
	for k := range ma.counter {
		delete(ma.counter, k)
	}
}

func New() *MetricAccumulator {

	gauge := make(map[string]float64)
	counter := make(map[string]int64)

	return &MetricAccumulator{gauge: gauge, counter: counter}
}

// оставляем
func (ms *MetricAccumulator) AddGaugeMetric(name string, value float64) {
	ms.Lock()
	defer ms.Unlock()

	ms.gauge[name] = value
}

// оставляем
func (ms *MetricAccumulator) AddCounterMetric(name string, value int64) {
	ms.Lock()
	defer ms.Unlock()

	if _, exists := ms.counter[name]; !exists {
		ms.counter[name] = value
	} else {
		ms.counter[name] += value
	}
}

// оставляем
func (ms *MetricAccumulator) GetAllCounterMetrics() map[string]int64 {
	return ms.counter
}

// оставляем
func (ms *MetricAccumulator) GetAllGaugeMetrics() map[string]float64 {
	return ms.gauge
}
