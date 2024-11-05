package metricservice

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/adettelle/go-metric-collector/internal/agent/config"
	m "github.com/adettelle/go-metric-collector/internal/agent/metrics"
	"github.com/adettelle/go-metric-collector/pkg/collections"

	"github.com/stretchr/testify/assert"
)

type MockMetricSender struct {
}

func (mms *MockMetricSender) SendMetricsChunk(id int, chunk []MetricRequest) error {
	return nil
}

func TestSendLoop(t *testing.T) {
	cfg := &config.Config{}

	ma := m.New()
	var wg sync.WaitGroup
	wg.Add(1)
	ms := NewMetricService(cfg, ma, &MockMetricSender{}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	go ms.SendLoop(ctx, 1, &wg)
	defer cancel()

	time.Sleep(2 * time.Second)
	// TODO create mock MetricSender and check
}

func TestRetrieveLoop(t *testing.T) {
	cfg := &config.Config{}

	ma := m.New()
	var wg sync.WaitGroup
	wg.Add(1)
	ms := NewMetricService(cfg, ma, &MockMetricSender{}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	go ms.RetrieveLoop(ctx, 1, &wg)
	defer cancel()

	time.Sleep(2 * time.Second)

	assert.NotEmpty(t, ma.GetAllGaugeMetrics())
	assert.NotEmpty(t, ma.GetAllCounterMetrics())
}

func TestAdditionalRetrieveLoop(t *testing.T) {
	cfg := &config.Config{}

	ma := m.New()
	var wg sync.WaitGroup
	wg.Add(1)
	ms := NewMetricService(cfg, ma, &MockMetricSender{}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	go ms.AdditionalRetrieveLoop(ctx, 1, &wg)
	defer cancel()

	time.Sleep(2 * time.Second)

	assert.NotEmpty(t, ma.GetAllGaugeMetrics())
	assert.Empty(t, ma.GetAllCounterMetrics())
}

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
		v1 = 1.1
		v2 = 2.2
		v3 = 3.3
	)
	metrics := []MetricRequest{
		{ID: "g1", MType: "gauge", Value: &v1},
		{ID: "g2", MType: "gauge", Value: &v2},
		{ID: "g3", MType: "gauge", Value: &v3},
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
