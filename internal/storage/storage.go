package storage

// IMemStorage Интрфейс с абстрактным функциями добавления, просмотра и удаления метрик в хранилище
type IMemStorage interface {
	SetMetric(metricType, metricName, metricValue string) error
	GetMetric(metricType, metricName string) (string, error)
	DeleteMetric(metricType, metricName string) error
}
