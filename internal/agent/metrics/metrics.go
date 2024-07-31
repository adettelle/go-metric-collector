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
	// sync.RWMutex
	gauge   *sync.Map // map[string]float64 // имя метрики: ее значение
	counter *sync.Map // map[string]int64
}

// Reset() обнуляет карты Gauge и Counter в структуре MemStorage
// метод применяется после отправки всех метрик
func (ma *MetricAccumulator) Reset() {
	ma.gauge.Range(func(key, value any) bool {
		ma.gauge.Delete(key)
		return true
	})
	ma.counter.Range(func(key, value any) bool {
		ma.counter.Delete(key)
		return true
	})
}

func New() *MetricAccumulator {
	gauge := &sync.Map{}
	counter := &sync.Map{}
	return &MetricAccumulator{gauge: gauge, counter: counter}
}

func (ma *MetricAccumulator) AddGaugeMetric(name string, value float64) {
	ma.gauge.Store(name, value)
}

func (ma *MetricAccumulator) AddCounterMetric(name string, value int64) {
	if v, exists := ma.counter.Load(name); !exists {
		ma.counter.Store(name, value)
	} else {
		ma.counter.Store(name, v.(int64)+value)
	}
}

func (ma *MetricAccumulator) GetAllCounterMetrics() map[string]int64 {
	result := make(map[string]int64)
	ma.counter.Range(func(key, value any) bool {
		result[key.(string)] = value.(int64)
		return true
	})
	return result
}

func (ma *MetricAccumulator) GetAllGaugeMetrics() map[string]float64 {
	result := make(map[string]float64)
	ma.gauge.Range(func(key, value any) bool {
		result[key.(string)] = value.(float64)
		return true
	})
	return result
}
