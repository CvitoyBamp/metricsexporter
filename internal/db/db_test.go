package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	user = "postgres"
	pass = "khjw7o9aJmCMVYJJ"
	url  = "db.zjqldcixgspktmukawkk.supabase.co"
	port = "5432"
	db   = "postgres"
)

type metric struct {
	metricName  string
	metricValue string
	metricType  string
}

func TestMemStorage_SetMetricDB(t *testing.T) {
	tests := []struct {
		testName   string
		db         *Database
		dataMetric metric
		wants      metric
	}{
		{
			testName: "Add gauge metric",
			db:       CreateDB(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, url, port, db)),
			dataMetric: metric{
				metricName:  "testGauge",
				metricType:  "gauge",
				metricValue: "1",
			},
			wants: metric{
				metricName:  "testGauge",
				metricType:  "gauge",
				metricValue: "1",
			},
		},
		{
			testName: "Add counter metric",
			db:       CreateDB(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, url, port, db)),
			dataMetric: metric{
				metricName:  "testCounter",
				metricType:  "counter",
				metricValue: "1",
			},
			wants: metric{
				metricName:  "testCounter",
				metricType:  "counter",
				metricValue: "1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			err := tt.db.SetMetricDB(tt.dataMetric.metricType, tt.dataMetric.metricName, tt.dataMetric.metricValue)
			require.NoError(t, err)
			metricG, errDB := tt.db.GetMetricDB(tt.dataMetric.metricType, tt.dataMetric.metricName)
			require.NoError(t, errDB)
			assert.Equal(t, tt.wants.metricValue, metricG)
		})
	}
}
