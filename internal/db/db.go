package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"log"
	"strconv"
	"time"
)

const (
	createGaugeTable = `CREATE TABLE IF NOT EXISTS gaugeMetrics(
        id        serial PRIMARY KEY,
        name      text NOT NULL,
        value     double precision NOT NULL,
        timestamp timestamp,
        UNIQUE (name))`
	createCounterTable = `CREATE TABLE IF NOT EXISTS counterMetrics(
        id        serial PRIMARY KEY,
        name      text NOT NULL,
        value     integer NOT NULL,
        timestamp timestamp,
        UNIQUE (name))`
	clearCounter = `DELETE FROM counterMetrics`
	getCount     = `WITH counter_count AS (SELECT COUNT(*) cc FROM counterMetrics),
                     gauge_count AS (SELECT COUNT(*) gc FROM gaugeMetrics)
        SELECT cc + gc AS sum_count
        FROM counter_count, gauge_count`
)

type Database struct {
	Conn *pgx.Conn
}

type Metrics struct {
	metricType  string
	metricName  string
	metricValue string
}

func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 1; i <= attempts; i++ {
		log.Printf("This is attempt number %d", i)
		err = f()
		if err != nil {
			log.Printf("error occured after attempt number %d: %s", i+1, err.Error())
			log.Println("sleeping for: ", sleep.String())
			time.Sleep(sleep)
			sleep += 2
			continue
		}
		break
	}
	return
}

func CreateDB(pgURL string) *Database {

	var db Database
	var errConn error

	attempts := 3
	duration := 1

	if pgURL != "" {

		ctx := context.Background()

		connConfig, err := pgx.ParseConfig(pgURL)

		if err != nil {
			log.Fatalf("Can't parse URL of PG DB, err: %s", err)
		}

		db.Conn, errConn = pgx.ConnectConfig(ctx, connConfig)

		if errConn != nil {
			if pgerrcode.IsConnectionException(errConn.Error()) {
				errR := Retry(attempts, time.Duration(duration), func() error {

					db.Conn, errConn = pgx.ConnectConfig(ctx, connConfig)

					return errConn
				})
				if errR != nil {
					log.Fatalf("Can't create connect to db, err: %s", errR)
				}
			}
			log.Fatalf("Can't create connect to db, err: %s", errConn)
		}

		_, err = db.Conn.Exec(context.Background(), createGaugeTable)
		if err != nil {
			log.Fatalf("Can't create table with gauge metrics, err: %s", err)
		}

		_, err = db.Conn.Exec(context.Background(), createCounterTable)
		if err != nil {
			log.Fatalf("Can't create table with counter metrics, err: %s", err)
		}

		_, err = db.Conn.Exec(context.Background(), clearCounter)
		if err != nil {
			log.Fatalf("Can't trunc counter table, err: %s", err)
		}
	}

	return &db
}

func (db Database) CheckConnectivity() error {
	return db.Conn.Ping(context.Background())
}

func (db Database) SetMetricDB(metricType, metricName, metricValue string) error {
	if metricType == "counter" {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("can't parse value to counter type (int64), error: %s", err)
		}
		_, err = db.Conn.Exec(context.Background(),
			`INSERT INTO counterMetrics (name, value, timestamp)
                 VALUES ($1, $2, $3)
                 ON CONFLICT (name) DO 
                 UPDATE SET value = (SELECT value FROM counterMetrics WHERE name = $1 ) + $2, timestamp = $3;`,
			metricName, value, time.Now())
		if err != nil {
			return err
		}
	} else if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("can't parse value to gauge type (float64), error: %s", err)
		}
		_, err = db.Conn.Exec(context.Background(),
			`INSERT INTO gaugeMetrics (name, value, timestamp)
                 VALUES ($1, $2, $3)
                 ON CONFLICT (name) DO UPDATE SET value = $2, timestamp = $3;`,
			metricName, value, time.Now())
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("don't know such type: %s", metricType)
	}

	return nil
}

func (db Database) GetMetricDB(metricType, metricName string) (string, error) {

	var metricValue string

	if metricType == "counter" {
		counterValue := db.Conn.QueryRow(context.Background(),
			`SELECT value FROM counterMetrics WHERE name = $1`, metricName)
		err := counterValue.Scan(&metricValue)
		if err != nil {
			return "", err
		}
	} else if metricType == "gauge" {
		gaugeValue := db.Conn.QueryRow(context.Background(),
			`SELECT value FROM gaugeMetrics WHERE name = $1`, metricName)
		err := gaugeValue.Scan(&metricValue)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("don't have metric's type %s in database", metricType)
	}

	return metricValue, nil
}

func (db Database) GetExistsMetricsDB() (map[string]string, error) {

	var metrics Metrics
	var count int

	rows := db.Conn.QueryRow(context.Background(), getCount)
	err := rows.Scan(&count)
	if err != nil {
		return nil, err
	}

	if count != 0 {
		metricsList := make(map[string]string, count)

		rowsGauge, errQG := db.Conn.Query(context.Background(), `SELECT name, value FROM gaugeMetrics`)
		if errQG != nil {
			return nil, errQG
		}

		for rowsGauge.Next() {
			errGScan := rowsGauge.Scan(&metrics.metricName, &metrics.metricValue)
			if errGScan != nil {
				return nil, errGScan
			}
			metricsList[metrics.metricName] = metrics.metricValue
		}

		rowsCounter, errCG := db.Conn.Query(context.Background(), `SELECT name, value FROM counterMetrics`)
		if errCG != nil {
			return nil, errCG
		}

		for rowsCounter.Next() {
			errCScan := rowsCounter.Scan(&metrics.metricName, &metrics.metricValue)
			if errCScan != nil {
				return nil, errCScan
			}
			metricsList[metrics.metricName] = metrics.metricValue
		}
		return metricsList, nil
	} else {
		return nil, fmt.Errorf("no metrics in storage for now")
	}
}
