// слой хранения: отвечает за хранение метрик (подразумевается два дериватива: положить и достать)
// аналог банковской ячейки (положить, достать)
package memstorage

import (
	"fmt"
	"log"
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

// MemStorage is used for storaging metrics
// MemStorage - это имплементация интерфейса Storage
type MemStorage struct {
	sync.RWMutex
	Gauge   map[string]float64 // имя метрики: ее значение
	Counter map[string]int64
	// если config.StoreInterval равен 0, то мы назначаем MemStorage FileName,
	// чтобы он мог синхронно писать изменения
	FileName string
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

	if ms.FileName != "" {
		err := WriteMetricsSnapshot(ms.FileName, ms)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func New(shouldRestore bool, storagePath string) (*MemStorage, error) {

	if shouldRestore {
		ms, err := ReadMetricsSnapshot(storagePath)
		if err != nil {
			return nil, err
		}
		ms.FileName = storagePath
		return ms, nil
	}

	gauge := make(map[string]float64)
	counter := make(map[string]int64)

	ms := &MemStorage{Gauge: gauge, Counter: counter, FileName: storagePath}

	// if storeInterval > 0 {
	// 	go StartSaveLoop(time.Second*time.Duration(storeInterval), storagePath, ms)
	// } else if storeInterval == 0 {
	// 	// если config.StoreInterval равен 0, то мы назначаем MemStorage FileName, чтобы
	// 	// он мог синхронно писать изменения
	// 	ms.FileName = storagePath
	// }

	return ms, nil
}

func (ms *MemStorage) GetGaugeMetric(name string) (float64, bool, error) {
	ms.RLock()
	defer ms.RUnlock()

	value, ok := ms.Gauge[name]

	return value, ok, nil
}

func (ms *MemStorage) GetCounterMetric(name string) (int64, bool, error) {
	ms.RLock()
	defer ms.RUnlock()

	value, ok := ms.Counter[name]

	return value, ok, nil
}

func (ms *MemStorage) AddGaugeMetric(name string, value float64) error {
	ms.Lock()
	defer ms.Unlock()

	ms.Gauge[name] = value

	if ms.FileName != "" {
		err := WriteMetricsSnapshot(ms.FileName, ms)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *MemStorage) AddCounterMetric(name string, value int64) error {
	ms.Lock()
	defer ms.Unlock()

	if _, exists := ms.Counter[name]; !exists {
		ms.Counter[name] = value
	} else {
		ms.Counter[name] += value
	}

	if ms.FileName != "" {
		err := WriteMetricsSnapshot(ms.FileName, ms)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ms *MemStorage) GetAllCounterMetrics() (map[string]int64, error) {
	return ms.Counter, nil
}

func (ms *MemStorage) GetAllGaugeMetrics() (map[string]float64, error) {
	return ms.Gauge, nil
}

// функция из структуры memStorage делает структуру AllMetrics
// перебираем ключи первой мэпы и перебираем ключи второй
func MemStorageToAllMetrics(ms *MemStorage) AllMetrics {
	var am AllMetrics

	for k, v := range ms.Gauge {
		am.AllMetrics = append(am.AllMetrics, Metric{ID: k, MType: "gauge", Value: &v})
	}
	for k, v := range ms.Counter {
		am.AllMetrics = append(am.AllMetrics, Metric{ID: k, MType: "counter", Delta: &v})
	}

	return am
}

func AllMetricsToMemStorage(am *AllMetrics) (*MemStorage, error) {
	ms, err := New(false, "")
	if err != nil {
		log.Fatal(err)
	}

	for _, metric := range am.AllMetrics {
		if metric.MType == "gauge" {
			ms.AddGaugeMetric(metric.ID, *metric.Value)
		} else if metric.MType == "counter" {
			ms.AddCounterMetric(metric.ID, *metric.Delta)
		} else {
			return nil, fmt.Errorf("unknown metric type: %s", metric.MType)
		}
	}

	return ms, nil
}

// отрабатывает завершение приложения (при штатном завершении работы)
// процесс финализации: объекты могут делать работу, пользоваться ресурсамии,
// и при заверщении работы (без работы с БД или с файлом), надо содержимое memStorage записать на диск (в файл)
func (ms *MemStorage) Finalize() error {
	log.Println("ms.FileName:", ms.FileName)
	return WriteMetricsSnapshot(ms.FileName, ms)
}
