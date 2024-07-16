package dbstorage

import (
	"context"
	"database/sql"
	"log"
)

// DBStorage - это имплементация интерфейса Storage
type DBStorage struct {
	Ctx context.Context
	DB  *sql.DB
}

func (dbstorage *DBStorage) GetGaugeMetric(name string) (float64, bool, error) {

	sqlStatement := "SELECT value FROM metric WHERE metric_type = 'gauge' and metric_id = $1"
	row := dbstorage.DB.QueryRowContext(dbstorage.Ctx, sqlStatement, name)

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

func (dbstorage *DBStorage) GetCounterMetric(name string) (int64, bool, error) {

	sqlStatement := "SELECT delta FROM metric WHERE metric_type = 'counter' and metric_id = $1"
	row := dbstorage.DB.QueryRowContext(dbstorage.Ctx, sqlStatement, name)

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

func (dbstorage *DBStorage) AddGaugeMetric(name string, value float64) error {
	log.Println("Writing to DB")

	tx, err := dbstorage.DB.Begin()
	if err != nil {
		return err
	}

	_, ok, err := dbstorage.GetGaugeMetric(name)
	if err != nil {
		return err
	}
	if !ok {
		sqlStatement := "insert into metric (metric_type, metric_id, value)" +
			"values ('gauge', $1, $2)"
		_, err := dbstorage.DB.ExecContext(dbstorage.Ctx, sqlStatement, name, value)
		if err != nil {
			log.Println("Error:", err)
			tx.Rollback()
			return err
		}
	} else {
		sqlStatement := "update metric set value = $1 where metric_type = 'gauge' and metric_id = $2"
		_, err := dbstorage.DB.ExecContext(dbstorage.Ctx, sqlStatement, value, name)
		if err != nil {
			log.Println("Error:", err)
			tx.Rollback()
			return err
		}
	}

	log.Println("Saved")
	return tx.Commit()
}

func (dbstorage *DBStorage) AddCounterMetric(name string, delta int64) error {
	log.Println("In AddCounterMetric")
	// начинаем транзакцию
	tx, err := dbstorage.DB.Begin()
	if err != nil {
		return err
	}

	oldDelta, ok, err := dbstorage.GetCounterMetric(name)
	if err != nil {
		log.Println("Err get:", err)
		return err
	}

	if !ok {
		sqlStatement := "insert into metric (metric_type, metric_id, delta)" +
			"values ('counter', $1, $2)"
		_, err := dbstorage.DB.ExecContext(dbstorage.Ctx, sqlStatement, name, delta)
		if err != nil {
			log.Println("Err ins:", err)
			// если ошибка, то откатываем изменения
			tx.Rollback()
			return err
		}
	} else {
		sqlStatement := "update metric set delta = $1 where metric_type = 'counter' and metric_id = $2"
		_, err := dbstorage.DB.ExecContext(dbstorage.Ctx, sqlStatement, oldDelta+delta, name)
		if err != nil {
			log.Println("Error upd:", err)
			// если ошибка, то откатываем изменения
			tx.Rollback()
			return err
		}
	}

	// завершаем транзакцию
	return tx.Commit()
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
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
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
	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return res, nil
}
