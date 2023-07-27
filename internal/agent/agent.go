package agent

import (
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/metrics"
	"log"
	"net/http"
	"runtime"
	"time"
)

type Agent struct {
	Client   *http.Client
	Endpoint string
	Metrics  *metrics.Metrics
}

func CreateAgent(endpoint string) *Agent {
	return &Agent{
		Client: &http.Client{
			Timeout: 1 * time.Second,
		},
		Endpoint: endpoint,
		Metrics: &metrics.Metrics{
			Gauge:   make(map[string]float64),
			Counter: make(map[string]int64),
		},
	}
}

func (a *Agent) PostMetric(metricType, metricName, metricValue string) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", a.Endpoint, metricType, metricName, metricValue)
	res, err := a.Client.Post(url, "text/plain", nil)
	if err != nil {
		log.Printf("metric %s with value %s was wasn't posted to %s\n", metricName, metricValue, url)
		return fmt.Errorf("can't POST to URL, err: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not equal 200: %v", res.StatusCode)
	}

	log.Printf("metric %s with value %s was successfully posted to %s\n", metricName, metricValue, url)

	return nil
}

func (a *Agent) PostMetrics() {
	for k, v := range a.Metrics.Gauge {
		err := a.PostMetric("gauge", k, fmt.Sprintf("%f", v))
		if err != nil {
			_ = fmt.Errorf("can't POST to URL, err: %v", err)
		}
	}
	for k, v := range a.Metrics.Counter {
		err := a.PostMetric("counter", k, fmt.Sprintf("%d", v))
		if err != nil {
			_ = fmt.Errorf("can't POST to URL, err: %v", err)
		}
	}
}

func (a *Agent) RunAgent(pollInterval, reportInterval int) {
	rI := time.NewTicker(time.Duration(reportInterval) * time.Second)
	pI := time.NewTicker(time.Duration(pollInterval) * time.Second)

	for {
		select {
		case <-pI.C:
			a.Metrics.MetricGenerator(runtime.MemStats{})
		case <-rI.C:
			a.PostMetrics()
		}
	}
}
