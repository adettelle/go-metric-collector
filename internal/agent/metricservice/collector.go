package metricservice

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"

	m "github.com/adettelle/go-metric-collector/internal/agent/metrics"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// RetrieveAllMetrics получает все метрики из пакета runtime
// и собирает дополнительные метрики (PollCount и RandomValue)
func RetrieveAllMetrics(metricAccumulator *m.MetricAccumulator) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metricAccumulator.AddCounterMetric("PollCount", 1)

	metricAccumulator.AddGaugeMetric("RandomValue", rand.Float64())

	metricAccumulator.AddGaugeMetric("Alloc", float64(m.Alloc))
	metricAccumulator.AddGaugeMetric("BuckHashSys", float64(m.BuckHashSys))
	metricAccumulator.AddGaugeMetric("Frees", float64(m.Frees))
	metricAccumulator.AddGaugeMetric("GCCPUFraction", m.GCCPUFraction)
	metricAccumulator.AddGaugeMetric("GCSys", float64(m.GCSys))
	metricAccumulator.AddGaugeMetric("HeapAlloc", float64(m.HeapAlloc))
	metricAccumulator.AddGaugeMetric("HeapIdle", float64(m.HeapIdle))
	metricAccumulator.AddGaugeMetric("HeapInuse", float64(m.HeapInuse))
	metricAccumulator.AddGaugeMetric("HeapObjects", float64(m.HeapObjects))
	metricAccumulator.AddGaugeMetric("HeapReleased", float64(m.HeapReleased))
	metricAccumulator.AddGaugeMetric("HeapSys", float64(m.HeapSys))
	metricAccumulator.AddGaugeMetric("LastGC", float64(m.LastGC))
	metricAccumulator.AddGaugeMetric("Lookups", float64(m.Lookups))
	metricAccumulator.AddGaugeMetric("MCacheInuse", float64(m.MCacheInuse))
	metricAccumulator.AddGaugeMetric("MCacheSys", float64(m.MCacheSys))
	metricAccumulator.AddGaugeMetric("MSpanInuse", float64(m.MSpanInuse))
	metricAccumulator.AddGaugeMetric("MSpanSys", float64(m.MSpanSys))
	metricAccumulator.AddGaugeMetric("Mallocs", float64(m.Mallocs))
	metricAccumulator.AddGaugeMetric("NextGC", float64(m.NextGC))
	metricAccumulator.AddGaugeMetric("NumForcedGC", float64(m.NumForcedGC))
	metricAccumulator.AddGaugeMetric("NumGC", float64(m.NumGC))
	metricAccumulator.AddGaugeMetric("OtherSys", float64(m.OtherSys))
	metricAccumulator.AddGaugeMetric("PauseTotalNs", float64(m.PauseTotalNs))
	metricAccumulator.AddGaugeMetric("StackInuse", float64(m.StackInuse))
	metricAccumulator.AddGaugeMetric("StackSys", float64(m.StackSys))
	metricAccumulator.AddGaugeMetric("Sys", float64(m.Sys))
	metricAccumulator.AddGaugeMetric("TotalAlloc", float64(m.TotalAlloc))
}

// retrieveAdditionalGaugeMetrics получает дополнительные метрики из пакета gopsutil
func retrieveAdditionalGaugeMetrics(metricAccumulator *m.MetricAccumulator) {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal(err)
	}
	cpu.Info()
	CPUutilizations, err := cpu.Percent(0, true)
	if err != nil {
		log.Fatal(err)
	}
	metricAccumulator.AddGaugeMetric("TotalMemory", float64(v.Total))
	metricAccumulator.AddGaugeMetric("FreeMemory", float64(v.Free))
	for i, CPUutilization := range CPUutilizations {
		metricAccumulator.AddGaugeMetric(fmt.Sprintf("CPUutilization%d", i+1), CPUutilization)
	}
}
