package db

type IDBStorage interface {
	SetMetricDB(metricType, metricName, metricValue string) error
	GetMetricDB(metricType, metricName string) (string, error)
	GetExistsMetricsDB() (map[string]string, error)
}
