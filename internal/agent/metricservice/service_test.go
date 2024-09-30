package metricservice

import (
	"testing"

	m "github.com/adettelle/go-metric-collector/internal/agent/metrics"
	"github.com/adettelle/go-metric-collector/pkg/collections"

	"github.com/stretchr/testify/assert"
)

func TestCollectAllMetrics(t *testing.T) {
	ma := m.New()
	ma.AddGaugeMetric("g1", 3.14)
	ma.AddCounterMetric("c1", 100)

	ms := &MetricService{
		metricAccumulator: ma,
	}

	metrics, err := ms.collectAllMetrics()
	assert.NoError(t, err)

	assert.Len(t, metrics, 2)

	// Check for correct gauge metric
	for _, metric := range metrics {
		if metric.ID == "g1" {
			assert.Equal(t, "gauge", metric.MType)
			assert.Equal(t, 3.14, *metric.Value)
		}

		if metric.ID == "c1" {
			assert.Equal(t, "counter", metric.MType)
			assert.Equal(t, int64(100), *metric.Delta)
		}
	}
}

// Test sends chunks of metrics
func TestSendMultipleMetrics(t *testing.T) {
	chunkSize := 2
	ma := m.New()
	ms := &MetricService{
		metricAccumulator: ma,
		ChunkSize:         chunkSize,
	}

	var (
		v1 float64 = 1.1
		v2 float64 = 2.2
		v3 float64 = 3.3
	)
	metrics := []MetricRequest{
		{ID: "g1", MType: "gauge", Value: &v1}, // new(float64)
		{ID: "g2", MType: "gauge", Value: &v2}, // new(float64)
		{ID: "g3", MType: "gauge", Value: &v3}, // new(float64)
	}

	chunksChannel := make(chan []MetricRequest, len(metrics))

	err := ms.sendMultipleMetrics(metrics, chunksChannel)
	assert.NoError(t, err)

	chunkCount := len(collections.RangeChunks(chunkSize, metrics))

	// Assert that chunks are properly sent
	for i := 0; i < chunkCount; i++ {
		chunk := <-chunksChannel
		assert.LessOrEqual(t, len(chunk), chunkSize)
	}
}
