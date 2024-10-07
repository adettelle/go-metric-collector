package memstorage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyMetricsToMemStorage(t *testing.T) {
	allMs := AllMetrics{}
	_, err := AllMetricsToMemStorage(&allMs)
	require.NoError(t, err)
}

func TestAllMetricsToMemStorage(t *testing.T) {
	d1 := int64(100)
	d2 := int64(200)
	d3 := int64(555)
	v1 := 1.1
	v2 := 10.1
	v3 := 1.23
	allMs := AllMetrics{
		AllMetrics: []Metric{
			{ID: "c1", MType: "counter", Delta: &d1},
			{ID: "c2", MType: "counter", Delta: &d2},
			{ID: "g1", MType: "gauge", Value: &v1},
			{ID: "g2", MType: "gauge", Value: &v2},
			// repeated metrics
			{ID: "c1", MType: "counter", Delta: &d3},
			{ID: "g1", MType: "gauge", Value: &v3},
		},
	}
	res, err := AllMetricsToMemStorage(&allMs)
	require.NoError(t, err)

	// получение всех counter метрик
	cMetrics, err := res.GetAllCounterMetrics()
	require.NoError(t, err)
	require.Equal(t, map[string]int64{"c1": 655, "c2": 200}, cMetrics)

	// получение всех gauge метрик
	gMetrics, err := res.GetAllGaugeMetrics()
	require.NoError(t, err)
	require.Equal(t, map[string]float64{"g1": 1.23, "g2": 10.1}, gMetrics)

}

func TestMemStorageToAllMetrics(t *testing.T) {
	d1 := int64(100)
	d2 := int64(200)
	d3 := int64(555)
	v1 := 1.1
	v2 := 10.1
	v3 := 1.23
	allMs := AllMetrics{
		AllMetrics: []Metric{
			{ID: "c1", MType: "counter", Delta: &d1},
			{ID: "c2", MType: "counter", Delta: &d2},
			{ID: "g1", MType: "gauge", Value: &v1},
			{ID: "g2", MType: "gauge", Value: &v2},
			// repeated metrics
			{ID: "c1", MType: "counter", Delta: &d3},
			{ID: "g1", MType: "gauge", Value: &v3},
		},
	}
	res, err := AllMetricsToMemStorage(&allMs)
	require.NoError(t, err)

	// sumD := d1 + d3
	// expectedMetrics := AllMetrics{
	// 	AllMetrics: []Metric{
	// 		{ID: "g1", MType: "gauge", Value: &v3},
	// 		{ID: "g2", MType: "gauge", Value: &v2},
	// 		{ID: "c1", MType: "counter", Delta: &sumD},
	// 		{ID: "c2", MType: "counter", Delta: &d2},
	// 	},
	// }

	res2 := MemStorageToAllMetrics(res)
	require.Equal(t, 4, len(res2.AllMetrics))
}

func TestIncorrectMetricType(t *testing.T) {
	d1 := int64(100)
	// d2 := int64(200)
	// d3 := int64(555)
	// v1 := 1.1
	// v2 := 10.1
	// v3 := 1.23
	allMs := AllMetrics{
		AllMetrics: []Metric{
			{ID: "c1", MType: "wrongType", Delta: &d1},
			// {ID: "c2", MType: "counter", Delta: &d2},
			// {ID: "g1", MType: "gauge", Value: &v1},
			// {ID: "g2", MType: "gauge", Value: &v2},
			// // repeated metrics
			// {ID: "c1", MType: "counter", Delta: &d3},
			// {ID: "g1", MType: "gauge", Value: &v3},
		},
	}
	_, err := AllMetricsToMemStorage(&allMs)
	require.Error(t, err)

	// получение всех counter метрик
	// cMetrics, err := res.GetAllCounterMetrics()
	// require.NoError(t, err)
	// require.Equal(t, map[string]int64{"c1": 655, "c2": 200}, cMetrics)

	// // получение всех gauge метрик
	// gMetrics, err := res.GetAllGaugeMetrics()
	// require.NoError(t, err)
	// require.Equal(t, map[string]float64{"g1": 1.23, "g2": 10.1}, gMetrics)
}
