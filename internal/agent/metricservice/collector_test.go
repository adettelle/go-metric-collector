package metricservice

import (
	"testing"

	m "github.com/adettelle/go-metric-collector/internal/agent/metrics"
	"github.com/stretchr/testify/require"
)

func TestRetrieveAllMetrics(t *testing.T) {
	accumulator := m.New()
	RetrieveAllMetrics(accumulator)

	counterMetrics := []string{}
	for k := range accumulator.GetAllCounterMetrics() {
		counterMetrics = append(counterMetrics, k)
	}

	require.ElementsMatch(t, []string{"PollCount"}, counterMetrics)

	gaugeMetrics := []string{}
	for k := range accumulator.GetAllGaugeMetrics() {
		gaugeMetrics = append(gaugeMetrics, k)
	}

	names := []string{
		"RandomValue",
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}
	require.ElementsMatch(t, names, gaugeMetrics)
}
