package dbstorage

import (
	"context"
	"database/sql"
	"log"
)

// DBStorage - это имплементация интерфейса Storage
type DBStorage struct {
	// Gauge   map[string]float64 // имя метрики: ее значение
	// Counter map[string]int64
	Ctx context.Context
	DB  *sql.DB
}

func (dbstorage *DBStorage) GetGaugeMetric(name string) (float64, bool, error) {

	sqlStatement := "SELECT value FROM metric WHERE metric_type = 'gauge' and metric_name = ?"
	row := dbstorage.DB.QueryRowContext(dbstorage.Ctx, sqlStatement, name)

	// переменная для чтения результата
	var val float64
	ok := true

	err := row.Scan(&val)
	if err != nil {
		ok = false
	}

	return val, ok, err
}

func (dbstorage *DBStorage) GetCounterMetric(name string) (int64, bool, error) {

	sqlStatement := "SELECT delta FROM metric WHERE metric_type = 'couter' and metric_name = ?"
	row := dbstorage.DB.QueryRowContext(dbstorage.Ctx, sqlStatement, name)

	// переменная для чтения результата
	var val int64
	ok := true

	err := row.Scan(&val)
	if err != nil {
		ok = false
	}

	return val, ok, err
}

func (dbstorage *DBStorage) AddGaugeMetric(name string, value float64) error {
	log.Println("Writing to DB")
	val, ok, err := dbstorage.GetGaugeMetric(name)
	if err != nil {
		return err
	}
	if !ok {
		sqlStatement := "insert into metric (metric_type, metric_name, value)" +
			"values ('gauge', $1, $2)"
		_, err := dbstorage.DB.ExecContext(dbstorage.Ctx, sqlStatement, name, value)
		if err != nil {
			log.Println("Error:", err)
			return err
		}
	} else {
		sqlStatement := "update metric set value = ? where metric_type = 'gauge' and metric_name = ?"
		_, err := dbstorage.DB.ExecContext(dbstorage.Ctx, sqlStatement, val, name)
		if err != nil {
			log.Println("Error:", err)
			return err
		}
	}

	log.Println("Saved")
	return nil
}

func (dbstorage *DBStorage) AddCounterMetric(name string, delta int64) error {

	d, ok, err := dbstorage.GetCounterMetric(name)
	if err != nil {
		return err
	}
	if !ok {
		sqlStatement := "insert into metric (metric_type, metric_name, delta)" +
			"values ('counter', $1, $2)"
		_, err := dbstorage.DB.ExecContext(dbstorage.Ctx, sqlStatement, name, delta)
		if err != nil {
			return err
		}
	} else {
		sqlStatement := "update metric set delta = ? where metric_type = 'counter' and metric_name = ?"
		_, err := dbstorage.DB.ExecContext(dbstorage.Ctx, sqlStatement, d, name)
		if err != nil {
			log.Println("Error:", err)
			return err
		}
	}

	return nil
}

func (dbstorage *DBStorage) GetAllCounterMetrics() (map[string]int64, error) {
	sqlStatement := "SELECT delta FROM metric WHERE metric_type = 'counter'"

	rows, err := dbstorage.DB.QueryContext(dbstorage.Ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]int64)
	for rows.Next() {
		var name string
		var d int64
		err := rows.Scan(&name, &d)
		if err != nil {
			return nil, err
		}
		res[name] = d
	}

	return res, nil
}

func (dbstorage *DBStorage) GetAllGaugeMetrics() (map[string]float64, error) {
	sqlStatement := "SELECT value FROM metric WHERE metric_type = 'gauge'"

	rows, err := dbstorage.DB.QueryContext(dbstorage.Ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]float64)
	for rows.Next() {
		var name string
		var v float64
		err := rows.Scan(&name, &v)
		if err != nil {
			return nil, err
		}
		res[name] = v
	}

	return res, nil
}
