package agent

import (
	"bytes"
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/db"
	"github.com/CvitoyBamp/metricsexporter/internal/json"
	"github.com/CvitoyBamp/metricsexporter/internal/metrics"
	"github.com/CvitoyBamp/metricsexporter/internal/middlewares"
	"log"
	"net/http"
	"runtime"
	"strconv"
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

func (a *Agent) PostMetricURL(metricType, metricName, metricValue string) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", a.Endpoint, metricType, metricName, metricValue)
	req, errReq := http.NewRequest(http.MethodPost, url, nil)
	if errReq != nil {
		return fmt.Errorf("can't create request, err: %v", errReq)
	}
	req.Close = true
	req.Header.Set("Content-Type", "text/plain")
	res, err := a.Client.Do(req)
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

func (a *Agent) PostMetricJSON(metricType, metricName, metricValue string) error {
	data, errJSON := json.Creator(metricValue, metricType, metricName)
	if errJSON != nil {
		log.Printf("can't convert body to json, err: %s", errJSON)
		return fmt.Errorf("can't convert body to json, err: %s", errJSON)
	}

	compressedData, errComp := middlewares.Compress(data)
	if errComp != nil {
		log.Printf("can't compress data, err: %s", errComp)
		return fmt.Errorf("can't compress data, err: %s", errComp)
	}

	url := fmt.Sprintf("http://%s/update/", a.Endpoint)

	req, errReq := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressedData))
	if errReq != nil {
		log.Printf("can't create request with body, err: %s", errReq)
		return fmt.Errorf("can't create request with body, err: %s", errReq)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := a.Client.Do(req)
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

func (a *Agent) PostMetricsBatch() error {
	data, errJSON := json.ListCreator(a.Metrics.Gauge, a.Metrics.Counter)
	if errJSON != nil {
		log.Printf("can't convert body to json, err: %s", errJSON)
		return fmt.Errorf("can't convert body to json, err: %s", errJSON)
	}

	compressedData, errComp := middlewares.Compress(data)
	if errComp != nil {
		log.Printf("can't compress data, err: %s", errComp)
		return fmt.Errorf("can't compress data, err: %s", errComp)
	}

	url := fmt.Sprintf("http://%s/updates/", a.Endpoint)

	req, errReq := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressedData))
	if errReq != nil {
		log.Printf("can't create request with body, err: %s", errReq)
		return fmt.Errorf("can't create request with body, err: %s", errReq)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := a.Client.Do(req)
	if err != nil {
		log.Printf("can't POST batch of metrics")
		return fmt.Errorf("can't POST to URL, err: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not equal 200: %v", res.StatusCode)
	}

	log.Printf("Batch of metrics was successfully posted.")

	return nil
}

func (a *Agent) PostMetrics(types string) error {
	a.Metrics.RLock()
	defer a.Metrics.RUnlock()

	if types == "batch" {
		err := a.PostMetricsBatch()
		if err != nil {
			return fmt.Errorf("can't POST to URL, err: %v", err)
		}
	}

	for k, v := range a.Metrics.Gauge {
		if types == "json" {
			err := a.PostMetricJSON("gauge", k, strconv.FormatFloat(v, 'f', -1, 64))
			if err != nil {
				return fmt.Errorf("can't POST to URL, err: %v", err)
			}
		}
		if types == "url" {
			err := a.PostMetricURL("gauge", k, strconv.FormatFloat(v, 'f', -1, 64))
			if err != nil {
				return fmt.Errorf("can't POST to URL, err: %v", err)
			}
		}
	}
	for k, v := range a.Metrics.Counter {
		if types == "json" {
			err := a.PostMetricJSON("counter", k, strconv.FormatInt(v, 10))
			if err != nil {
				return fmt.Errorf("can't POST to URL, err: %v", err)
			}
		}
		if types == "url" {
			err := a.PostMetricURL("counter", k, strconv.FormatInt(v, 10))
			if err != nil {
				return fmt.Errorf("can't POST to URL, err: %v", err)
			}
		}
	}
	return nil
}

func (a *Agent) RunAgent(pollInterval, reportInterval int) {
	rI := time.NewTicker(time.Duration(reportInterval) * time.Second)
	pI := time.NewTicker(time.Duration(pollInterval) * time.Second)
	attempts := 3
	duration := 1

	for {
		select {
		case <-pI.C:
			a.Metrics.MetricGenerator(runtime.MemStats{})
		case <-rI.C:
			err := db.Retry(attempts, time.Duration(duration), func() error {
				err := a.PostMetrics("url")
				return err
			})
			if err != nil {
				log.Print(err)
			}
		}
	}
}
