package memstorage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	// "go.uber.org/zap"
)

// Test WriteMetricsSnapshot writes the metrics data to a file and reads it back to confirm accuracy
func TestWriteMetricsSnapshot(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "metrics.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // Clean up the file after the test

	ms, err := New(false, tmpFile.Name())
	assert.NoError(t, err)

	err = ms.AddGaugeMetric("g1", 0.75)
	assert.NoError(t, err)
	err = ms.AddGaugeMetric("g2", 0.123)
	assert.NoError(t, err)
	err = ms.AddCounterMetric("c1", 1)
	assert.NoError(t, err)
	err = ms.AddCounterMetric("c2", 2)
	assert.NoError(t, err)
	allmetricsFromMs := MemStorageToAllMetrics(ms)

	err = WriteMetricsSnapshot(tmpFile.Name(), ms)
	assert.NoError(t, err)

	data, err := ReadMetricsSnapshot(tmpFile.Name())
	assert.NoError(t, err)
	allmetricsFromData := MemStorageToAllMetrics(data)

	assert.ElementsMatch(t, allmetricsFromMs.AllMetrics, allmetricsFromData.AllMetrics)
}

func TestReadMetricsSnapshotInexistent(t *testing.T) {
	_, err := ReadMetricsSnapshot("idontexist.txt")
	assert.Error(t, err)
}

/*
// Test ReadMetricsSnapshot reads the metrics data from a file and checks accuracy
func TestReadMetricsSnapshot(t *testing.T) {
	// metricsData := AllMetrics{Metrics: map[string]float64{"cpu": 0.85, "memory": 0.45}}
	tmpFile, err := os.CreateTemp("", "metrics_read.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // Clean up the file after the test

	ms, err := New(false, tmpFile.Name())
	assert.NoError(t, err)

	err = ms.AddGaugeMetric("g1", 0.75)
	assert.NoError(t, err)
	err = ms.AddGaugeMetric("g2", 0.123)
	assert.NoError(t, err)

	gMetrics, err := ms.GetAllGaugeMetrics()
	assert.NoError(t, err)
	memStorageFromMs := MemStorageToAllMetrics(ms)

	// Write initial data to temp file
	data, err := json.Marshal(gMetrics) // metricsData
	assert.NoError(t, err)
	_, err = tmpFile.Write(data)
	assert.NoError(t, err)
	tmpFile.Close()

	// Read the data back and verify
	msRes, err := ReadMetricsSnapshot(tmpFile.Name())
	assert.NoError(t, err)
	msResAllMetrics := MemStorageToAllMetrics(msRes)
	// gMetricsRes, err := ms.GetAllGaugeMetrics()
	// assert.NoError(t, err)
	// assert.Equal(t, metricsData.Metrics, ms.Metrics)
	assert.Equal(t, msResAllMetrics, memStorageFromMs)
}
*/
/*
// Test StartSaveLoop simulates the save loop by controlling the ticker
func TestStartSaveLoop(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "loop_metrics.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // Clean up the file after the test

	ms, err := New(false, tmpFile.Name())
	assert.NoError(t, err)

	err = ms.AddGaugeMetric("g1", 0.75)
	assert.NoError(t, err)
	err = ms.AddGaugeMetric("g2", 0.123)
	assert.NoError(t, err)

	// Set up a short duration and a channel to exit the loop
	storeInterval := 10 * time.Millisecond
	stop := make(chan struct{})
	go func() {
		StartSaveLoop(storeInterval, tmpFile.Name(), ms)
		stop <- struct{}{}
	}()

	// Wait briefly for the loop to write at least once
	time.Sleep(30 * time.Millisecond)
	close(stop)

	// Verify that file was written to at least once
	data, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	var result AllMetrics
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, , )
}
*/
