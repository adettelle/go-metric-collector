package dbstorage

import (
	"context"
	"database/sql"
	"log"

	"github.com/adettelle/go-metric-collector/internal/api"
)

var (
	_ api.Storager = (*DBStorage)(nil)
)

// DBStorage - это имплементация (или реализация) интерфейса Storage
type DBStorage struct {
	Ctx context.Context
	DB  *sql.DB
}

func (s *DBStorage) GetGaugeMetric(name string) (float64, bool, error) {

	sqlStatement := "SELECT value FROM metric WHERE metric_type = 'gauge' and metric_id = $1"
	row := s.DB.QueryRowContext(s.Ctx, sqlStatement, name)

	// переменная для чтения результата
	var val float64

	err := row.Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}

	return val, true, err
}

func (s *DBStorage) GetCounterMetric(name string) (int64, bool, error) {

	sqlStatement := "SELECT delta FROM metric WHERE metric_type = 'counter' and metric_id = $1"
	row := s.DB.QueryRowContext(s.Ctx, sqlStatement, name)

	// переменная для чтения результата
	var val int64

	err := row.Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}

	return val, true, err
}

func (s *DBStorage) AddGaugeMetric(name string, value float64) error {
	log.Println("Writing to DB")

	sqlStatement := `insert into metric (metric_type, metric_id, value) 
		values ('gauge', $1, $2) on conflict (metric_id, metric_type) do update set value = $2`

	_, err := s.DB.ExecContext(s.Ctx, sqlStatement, name, value)
	if err != nil {
		log.Println("error in updating gauge metric:", err)
		return err
	}

	log.Println("Saved")
	return nil
}

func (s *DBStorage) AddCounterMetric(name string, delta int64) error {
	log.Println("In AddCounterMetric")

	sqlStatement := `insert into metric (metric_type, metric_id, delta)
		values ('counter', $1, $2)
		on conflict (metric_id, metric_type) do update set
		delta = (select delta from metric where metric_type = 'counter' and metric_id = $1) + $2`

	_, err := s.DB.ExecContext(s.Ctx, sqlStatement, name, delta)
	if err != nil {
		log.Println("error in updating counter metric:", err)
		return err
	}

	return nil
}

func (s *DBStorage) GetAllCounterMetrics() (map[string]int64, error) {
	sqlStatement := "SELECT metric_id, delta FROM metric WHERE metric_type = 'counter'"

	rows, err := s.DB.QueryContext(s.Ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]int64)
	for rows.Next() {
		var name string
		var d int64
		if err = rows.Scan(&name, &d); err != nil {
			return nil, err
		}
		res[name] = d
	}
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *DBStorage) GetAllGaugeMetrics() (map[string]float64, error) {
	var err error

	sqlStatement := "SELECT metric_id, value FROM metric WHERE metric_type = 'gauge'"

	rows, err := s.DB.QueryContext(s.Ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]float64)
	for rows.Next() {
		var name string
		var v float64
		if err = rows.Scan(&name, &v); err != nil {
			return nil, err
		}

		res[name] = v
	}
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *DBStorage) Finalize() error {
	return nil
}
