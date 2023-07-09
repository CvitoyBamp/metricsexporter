package agent

import (
	"context"
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/metrics"
	"io"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type Agent struct {
	Client   *http.Client
	Endpoint string
	MemStats *runtime.MemStats
	Metrics  *metrics.Metrics
}

func CreateAgent(endpoint string) *Agent {
	return &Agent{
		Client: &http.Client{
			Timeout: 1 * time.Second,
		},
		Endpoint: endpoint,
		MemStats: &runtime.MemStats{},
		Metrics: &metrics.Metrics{
			Gauge:   make(map[string]float64),
			Counter: make(map[string]int64),
		},
	}
}

func (a *Agent) postMetric(metricType, metricName, metricValue string) error {
	url := fmt.Sprintf("%s/update/%s/%s/%s", a.Endpoint, metricType, metricName, metricValue)
	res, err := a.Client.Post(url, "text/plain", nil)
	fmt.Printf("metric %s with value %s was successfully posted to %s\n", metricName, metricValue, url)
	if err != nil {
		return fmt.Errorf("can't POST to URL, err: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal("fatal")
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not equal 200: %v", res.StatusCode)
	}

	return fmt.Errorf("can't POST metric to %s", url)
}

func (a *Agent) pollMetrics(ctx context.Context, w *sync.WaitGroup, pollInterval int) {
	defer w.Done()
	pI := time.NewTicker(time.Duration(pollInterval))
	defer pI.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pI.C:
			a.Metrics.Lock()
			a.Metrics.MetricGenerator(*a.MemStats)
			a.Metrics.Unlock()
		}
	}
}

func (a *Agent) postMetrics(ctx context.Context, w *sync.WaitGroup, reportInterval int) {
	defer w.Done()
	rI := time.NewTicker(time.Duration(reportInterval))
	defer rI.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rI.C:
			a.Metrics.Lock()
			for k, v := range a.Metrics.Gauge {
				err := a.postMetric("gauge", k, fmt.Sprintf("%f", v))
				if err != nil {
					_ = fmt.Errorf("can't POST to URL, err: %v", err)
				}
			}
			for k, v := range a.Metrics.Counter {
				err := a.postMetric("counter", k, fmt.Sprintf("%d", v))
				if err != nil {
					_ = fmt.Errorf("can't POST to URL, err: %v", err)
				}
			}
			a.Metrics.Unlock()
		}
	}
}

func (a *Agent) RunAgent(pollInterval, reportInterval int) {
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go a.pollMetrics(context.TODO(), wg, pollInterval)
	go a.postMetrics(context.TODO(), wg, reportInterval)
	wg.Wait()
}
