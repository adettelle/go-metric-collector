package storage

import "sync"

// MemStorage is used for storaging metrics
// MemStorage - это имплементация интерфейса Storage
type MemStorage struct {
	sync.RWMutex
	Gauge   map[string]float64
	Counter map[string]int64
}

// Reset() обнуляет карты Gauge и Counter в структуре MemStorage
// метод применяется после отправки всех метрик
func (ms *MemStorage) Reset() {
	ms.Lock()
	defer ms.Unlock()

	for k := range ms.Gauge {
		delete(ms.Gauge, k)
	}
	for k := range ms.Counter {
		delete(ms.Counter, k)
	}
}

func NewMemStorage() *MemStorage {
	gauge := make(map[string]float64)
	counter := make(map[string]int64)
	return &MemStorage{Gauge: gauge, Counter: counter}
}

func (ms *MemStorage) GetGaugeMetric(name string) (float64, bool) {
	ms.RLock()
	defer ms.RUnlock()

	value, ok := ms.Gauge[name]
	return value, ok
}

func (ms *MemStorage) GetCounterMetric(name string) (int64, bool) {
	ms.RLock()
	defer ms.RUnlock()

	value, ok := ms.Counter[name]
	return value, ok
}

func (ms *MemStorage) AddGaugeMetric(name string, value float64) {
	ms.Lock()
	defer ms.Unlock()

	ms.Gauge[name] = value
}

func (ms *MemStorage) AddCounterMetric(name string, value int64) {
	ms.Lock()
	defer ms.Unlock()

	if _, exists := ms.Counter[name]; !exists {
		ms.Counter[name] = value
	} else {
		ms.Counter[name] += value
	}
}

// интерфейс для взаимодействия с хранилищем MemStorage и другими хранилищами, напрмер, fileStorage
type Storage interface {
	GetGaugeMetric(name string) (float64, bool)
	GetCounterMetric(name string) (int64, bool)
	AddGaugeMetric(name string, value float64)
	AddCounterMetric(name string, value int64)
}
