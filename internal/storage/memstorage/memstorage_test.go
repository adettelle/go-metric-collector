package memstorage_test

import (
	"fmt"
	"log"
	"sort"
	"testing"

	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
	"github.com/stretchr/testify/require"
)

func ExampleMemStorage_AddCounterMetric() {
	ms, err := memstorage.New(false, "")
	if err != nil {
		log.Fatal(err)
	}

	name := "someMetric"
	var (
		value1 int64 = 525
		value2 int64 = 100
	)

	// записали метрику в хранилище
	ms.AddCounterMetric(name, value1)
	// проверка наличия метрики в map
	val1, ok, err := ms.GetCounterMetric(name)
	if ok {
		fmt.Println(val1)
	}

	ms.AddCounterMetric(name, value2)
	val2, ok, err := ms.GetCounterMetric(name)
	if ok {
		fmt.Println(val2)
	}

	// Output:
	// 525
	// 625
}

func ExampleMemStorage_AddGaugeMetric() {
	ms, err := memstorage.New(false, "")
	if err != nil {
		log.Fatal(err)
	}

	name := "someGaugeMetric"

	value1 := 111.222
	value2 := 100.555

	// записали метрику в хранилище
	ms.AddGaugeMetric(name, value1)
	// проверка наличия метрики в map
	val1, ok, err := ms.GetGaugeMetric(name)
	if ok {
		fmt.Println(val1)
	}

	ms.AddGaugeMetric(name, value2)
	val2, ok, err := ms.GetGaugeMetric(name)
	if ok {
		fmt.Println(val2)
	}

	// Output:
	// 111.222
	// 100.555
}

func ExampleMemStorage_GetAllCounterMetrics() {
	ms, err := memstorage.New(false, "")
	if err != nil {
		log.Fatal(err)
	}

	metrics := map[string]int64{
		"c1": 525,
		"c2": 100,
	}

	for k, v := range metrics {
		ms.AddCounterMetric(k, v)
	}

	checkmetrics, err := ms.GetAllCounterMetrics()
	if err != nil {
		log.Fatal(err)
	}

	names := make([]string, 0, len(metrics))
	for name := range metrics {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if metrics[name] == checkmetrics[name] {
			fmt.Println(metrics[name])
		} else {
			log.Fatal("Not equal")
		}
	}
	// Output:
	// 525
	// 100
}

func TestReset(t *testing.T) {
	ms, err := memstorage.New(false, "")
	require.NoError(t, err)

	err = ms.AddCounterMetric("m1", 1)
	require.NoError(t, err)

	err = ms.AddGaugeMetric("g1", 1.1)
	require.NoError(t, err)

	ms.Reset()

	cMetrics, err := ms.GetAllCounterMetrics()
	require.NoError(t, err)
	require.Empty(t, cMetrics)

	gMetrics, err := ms.GetAllGaugeMetrics()
	require.NoError(t, err)
	require.Empty(t, gMetrics)

}
