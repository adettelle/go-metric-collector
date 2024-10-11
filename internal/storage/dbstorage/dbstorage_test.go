package dbstorage

import (
	"context"
	"testing"

	database "github.com/adettelle/go-metric-collector/internal/db"
	"github.com/adettelle/go-metric-collector/internal/migrator"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

func TestDBStorage(t *testing.T) {
	dbParams := "host=localhost port=9999 user=postgres password=123456 dbname=test_db sslmode=disable"

	err := migrator.ApplyMigrations(dbParams)
	require.NoError(t, err)

	defer func() {
		if err = migrator.ResetMigrations(dbParams); err != nil {
			t.Fatal(err)
		}
	}()

	db, err := database.NewDBConnection(dbParams).Connect()
	require.NoError(t, err)

	sDB := &DBStorage{
		Ctx: context.Background(),
		DB:  db,
	}

	cm1Name := uuid.NewString()[:30]
	cm2Name := uuid.NewString()[:30]

	err = sDB.AddCounterMetric(cm1Name, 100)
	require.NoError(t, err)

	// получение существующей метрики
	cm1Value, ok, err := sDB.GetCounterMetric(cm1Name)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, int64(100), cm1Value)

	// получение несуществующей метрики
	_, ok, err = sDB.GetCounterMetric("inexistentCounter")
	require.NoError(t, err)
	require.False(t, ok)

	// добавление той же (существующей) counter метрики, должно сохранить сумму двух значений
	err = sDB.AddCounterMetric(cm1Name, 111)
	require.NoError(t, err)

	// получение существующей метрики
	cm1Value, ok, err = sDB.GetCounterMetric(cm1Name)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, int64(211), cm1Value)

	// добавление другой counter метрики
	err = sDB.AddCounterMetric(cm2Name, 200)
	require.NoError(t, err)

	// получение всех counter метрик
	cMetrics, err := sDB.GetAllCounterMetrics()
	require.NoError(t, err)
	require.Equal(t, map[string]int64{cm1Name: 211, cm2Name: 200}, cMetrics)

	// ------
	gm1Name := uuid.NewString()[:30]
	gm2Name := uuid.NewString()[:30]

	err = sDB.AddGaugeMetric(gm1Name, 1.1)
	require.NoError(t, err)

	// получение существующей метрики
	gm1Value, ok, err := sDB.GetGaugeMetric(gm1Name)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 1.1, gm1Value)

	// получение несуществующей метрики
	_, ok, err = sDB.GetGaugeMetric("inexistentGauge")
	require.NoError(t, err)
	require.False(t, ok)

	// добавление той же (существующей) gauge метрики, должно перезаписывать значение
	err = sDB.AddGaugeMetric(gm1Name, 2.2)
	require.NoError(t, err)

	// получение существующей метрики
	gm1Value, ok, err = sDB.GetGaugeMetric(gm1Name)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, 2.2, gm1Value)

	// добавление другой counter метрики
	err = sDB.AddGaugeMetric(gm2Name, 22.222)
	require.NoError(t, err)

	// получение всех gauge метрик
	gMetrics, err := sDB.GetAllGaugeMetrics()
	require.NoError(t, err)
	require.Equal(t, map[string]float64{gm1Name: 2.2, gm2Name: 22.222}, gMetrics)

	err = sDB.Finalize()
	require.NoError(t, err)
}

func TestDBStorageNoDB(t *testing.T) {
	dbParams := "host=localhost port=9999 user=postgres password=123456 dbname=test_db sslmode=disable"

	db, err := database.NewDBConnection(dbParams).Connect()
	require.NoError(t, err)

	sDB := &DBStorage{
		Ctx: context.Background(),
		DB:  db,
	}

	cm1Name := uuid.NewString()[:30]
	cm2Name := uuid.NewString()[:30]

	err = sDB.AddCounterMetric(cm1Name, 100)
	require.Error(t, err)

	// получение существующей метрики
	_, _, err = sDB.GetCounterMetric(cm1Name)
	require.Error(t, err)

	// получение несуществующей метрики
	_, _, err = sDB.GetCounterMetric("inexistentCounter")
	require.Error(t, err)

	// добавление той же (существующей) counter метрики, должно сохранить сумму двух значений
	err = sDB.AddCounterMetric(cm1Name, 111)
	require.Error(t, err)

	// получение существующей метрики
	_, _, err = sDB.GetCounterMetric(cm1Name)
	require.Error(t, err)

	// добавление другой counter метрики
	err = sDB.AddCounterMetric(cm2Name, 200)
	require.Error(t, err)

	// получение всех counter метрик
	_, err = sDB.GetAllCounterMetrics()
	require.Error(t, err)

	// ------
	gm1Name := uuid.NewString()[:30]
	gm2Name := uuid.NewString()[:30]

	err = sDB.AddGaugeMetric(gm1Name, 1.1)
	require.Error(t, err)

	// получение существующей метрики
	_, _, err = sDB.GetGaugeMetric(gm1Name)
	require.Error(t, err)

	// получение несуществующей метрики
	_, _, err = sDB.GetGaugeMetric("inexistentGauge")
	require.Error(t, err)

	// добавление той же (существующей) gauge метрики, должно перезаписывать значение
	err = sDB.AddGaugeMetric(gm1Name, 2.2)
	require.Error(t, err)

	// получение существующей метрики
	_, _, err = sDB.GetGaugeMetric(gm1Name)
	require.Error(t, err)

	// добавление другой counter метрики
	err = sDB.AddGaugeMetric(gm2Name, 22.222)
	require.Error(t, err)

	// получение всех gauge метрик
	_, err = sDB.GetAllGaugeMetrics()
	require.Error(t, err)

	err = sDB.Finalize()
	require.NoError(t, err)
}
