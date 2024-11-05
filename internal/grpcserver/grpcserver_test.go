package grpcserver

import (
	"testing"
	"time"

	"github.com/adettelle/go-metric-collector/internal/agent/metricservice"
	"github.com/adettelle/go-metric-collector/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGrpcServer(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	// создаём объект-заглушку
	m := mocks.NewMockStorager(ctrl)

	go func() {
		_ = StartServer(m, "3333")
	}()

	time.Sleep(100 * time.Millisecond)
	sender := metricservice.NewGrpcSender("localhost:3333")

	m.EXPECT().AddCounterMetric(gomock.Any(), gomock.Any())
	m.EXPECT().AddGaugeMetric(gomock.Any(), gomock.Any())

	delta := int64(1)
	value := 11.22

	err := sender.SendMetricsChunk(1, []metricservice.MetricRequest{
		{ID: "m1", MType: "counter", Delta: &delta},
		{ID: "m2", MType: "gauge", Value: &value},
	})
	require.NoError(t, err)

}
