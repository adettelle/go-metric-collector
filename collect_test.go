package metricservice

import (
	"testing"

	"github.com/adettelle/go-metric-collector/internal/agent/metrics"
	"github.com/adettelle/go-metric-collector/internal/agent/metricservice"
)

func BenchmarkRetrieveAllMetrics(b *testing.B) {
	metricAccumulator := metrics.New()
	for i := 0; i < b.N; i++ {
		metricservice.RetrieveAllMetrics(metricAccumulator)
	}
}
