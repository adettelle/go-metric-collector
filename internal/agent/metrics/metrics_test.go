package metrics

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricAccumulator(t *testing.T) {
	ma := New()

	assert.NotNil(t, ma)
	assert.NotNil(t, ma.gauge)
	assert.NotNil(t, ma.counter)
}

func TestAddCounterMetric(t *testing.T) {
	name := "C1"
	var delta int64 = 525

	ma := New()
	allMs := ma.GetAllCounterMetrics()
	require.Equal(t, allMs, map[string]int64{}) // require.Nill(t, allMs)

	lenBeforeAdding := len(allMs)
	ma.AddCounterMetric(name, delta)
	allMsAfter := ma.GetAllCounterMetrics()
	lenAfterAdding := len(allMsAfter)
	assert.NotEqual(t, lenBeforeAdding, lenAfterAdding)

	// проверка наличия метрики в map
	cMetrics1 := ma.GetAllCounterMetrics()
	val1, ok := cMetrics1[name]
	assert.True(t, ok)
	assert.Equal(t, val1, delta)

	// проверка добавление уже сущ-ей метрики
	ma.AddCounterMetric(name, delta)
	cMetrics2 := ma.GetAllCounterMetrics()
	val2, ok := cMetrics2[name]
	assert.True(t, ok)
	assert.Equal(t, val2, delta+delta)

	// проверка получения несущ-ей метрики
	unrealName := "UnrealMetric"
	var zero int64 = 0
	cMetrics3 := ma.GetAllCounterMetrics()
	val3, ok := cMetrics3[unrealName]
	assert.False(t, ok)
	assert.Equal(t, val3, zero)
}

func TestAddGaugeMetric(t *testing.T) {
	name := "G1"
	var value = 1.123

	ma := New()
	allMs := ma.GetAllGaugeMetrics()
	require.Equal(t, allMs, map[string]float64{}) // require.Nill(t, allMs)

	lenBeforeAdding := len(allMs)
	ma.AddGaugeMetric(name, value)
	allMsAfter := ma.GetAllGaugeMetrics()
	lenAfterAdding := len(allMsAfter)
	assert.NotEqual(t, lenBeforeAdding, lenAfterAdding)

	// проверка наличия метрики в map
	gMetrics1 := ma.GetAllGaugeMetrics()
	val1, ok := gMetrics1[name]
	assert.True(t, ok)
	assert.Equal(t, val1, value)

	// проверка добавление уже сущ-ей метрики
	var newValue = 222.666
	ma.AddGaugeMetric(name, newValue)
	gMetrics2 := ma.GetAllGaugeMetrics()
	val2, ok := gMetrics2[name]
	assert.True(t, ok)
	assert.Equal(t, val2, newValue)

	// проверка получения несущ-ей метрики
	unrealName := "UnrealMetric"

	gMetrics3 := ma.GetAllGaugeMetrics()
	val3, ok := gMetrics3[unrealName]
	assert.False(t, ok)
	assert.Equal(t, val3, 0.0)
}

func TestReset(t *testing.T) {
	ma := New()
	ma.AddGaugeMetric("g1", 22.5)
	ma.AddCounterMetric("c1", 5)

	// Reset the metrics
	ma.Reset()

	// Check if the gauge and counter maps are empty after reset
	require.Equal(t, ma.GetAllGaugeMetrics(), map[string]float64{})
	require.Equal(t, ma.GetAllCounterMetrics(), map[string]int64{})
}

func TestConcurrentAccess(t *testing.T) {
	ma := New()

	var wg sync.WaitGroup

	// Add metrics concurrently
	for i := 0; i < 1000; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			ma.AddGaugeMetric("g1", float64(i))
		}(i)

		go func(i int) {
			defer wg.Done()
			ma.AddCounterMetric("c1", int64(i))
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check that metrics exist (just a sample check)
	assert.NotEqual(t, len(ma.GetAllGaugeMetrics()), 0)
	assert.NotEqual(t, len(ma.GetAllCounterMetrics()), 0)
}

func TestGetAllGaugeMetrics(t *testing.T) {
	ma := New()
	ma.AddGaugeMetric("g1", 0.75)
	ma.AddGaugeMetric("g2", 0.65)

	expected := map[string]float64{
		"g1": 0.75,
		"g2": 0.65,
	}

	gaugeMetrics := ma.GetAllGaugeMetrics()
	assert.True(t, reflect.DeepEqual(gaugeMetrics, expected))
}

func TestGetAllCounterMetrics(t *testing.T) {
	ma := New()
	ma.AddCounterMetric("c1", 10)
	ma.AddCounterMetric("c2", 2)

	expected := map[string]int64{
		"c1": 10,
		"c2": 2,
	}

	counterMetrics := ma.GetAllCounterMetrics()
	assert.True(t, reflect.DeepEqual(counterMetrics, expected))
}
