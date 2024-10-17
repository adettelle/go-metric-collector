// Storage layer: responsible for storing metrics (implies two derivatives: put and take out)
// analogous to a safety deposit box (put, take out).
package metrics

import (
	"sync"
)

type Metric struct {
	Delta *int64   `json:"delta,omitempty"` // metric's value when metric type is counter
	Value *float64 `json:"value,omitempty"` // metric's value when metric type is gauge
	ID    string   `json:"id"`              // metric's name
	MType string   `json:"type"`            // parameter that takes gauge or counter value
}

type AllMetrics struct {
	AllMetrics []Metric `json:"metrics"`
}

// MetricAccumulator is used for storaging metrics.
type MetricAccumulator struct {
	gauge   *sync.Map // map[string]float64 // имя метрики: ее значение
	counter *sync.Map // map[string]int64
}

func New() *MetricAccumulator {
	gauge := &sync.Map{}
	counter := &sync.Map{}
	return &MetricAccumulator{gauge: gauge, counter: counter}
}

// Reset() resets the Gauge and Counter maps in the MemStorage structure,
// the method is applied after sending all metrics.
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
