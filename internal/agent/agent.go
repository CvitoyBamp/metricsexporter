package agent

import (
	"bytes"
	"context"
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/crypto"
	"github.com/CvitoyBamp/metricsexporter/internal/db"
	"github.com/CvitoyBamp/metricsexporter/internal/json"
	"github.com/CvitoyBamp/metricsexporter/internal/metrics"
	"github.com/CvitoyBamp/metricsexporter/internal/middlewares"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

type Agent struct {
	Client   *http.Client
	Endpoint string
	Metrics  *metrics.Metrics
	Config   *Config
}

func CreateAgent(cfg Config) *Agent {
	return &Agent{
		Client: &http.Client{
			Timeout: 1 * time.Second,
		},
		Endpoint: cfg.Address,
		Metrics: &metrics.Metrics{
			Gauge:   make(map[string]float64),
			Counter: make(map[string]int64),
		},
		Config: &cfg,
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

	if a.Config.Key != "" {
		req.Header.Set("HashSHA256", fmt.Sprintf("%x", crypto.CreateHash(compressedData, a.Config.Key)))
	}

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

	if a.Config.Key != "" {
		req.Header.Set("HashSHA256", fmt.Sprintf("%x", crypto.CreateHash(compressedData, a.Config.Key)))
	}

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

func (a *Agent) PostMetrics(types string, metrics <-chan json.Metrics, done chan<- bool) error {

	if types == "batch" {
		err := a.PostMetricsBatch()
		if err != nil {
			return fmt.Errorf("can't POST to URL, err: %v", err)
		}
	}

	for metric := range metrics {

		if types == "json" {
			if metric.MType == "gauge" {
				err := a.PostMetricJSON(metric.MType, metric.ID, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
				if err != nil {
					return fmt.Errorf("can't POST to URL, err: %v", err)
				}
			}

			if metric.MType == "counter" {
				err := a.PostMetricJSON(metric.MType, metric.ID, strconv.FormatInt(*metric.Delta, 10))
				if err != nil {
					return fmt.Errorf("can't POST to URL, err: %v", err)
				}
			}

		}

		if types == "url" {
			if metric.MType == "gauge" {
				err := a.PostMetricURL(metric.MType, metric.ID, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
				if err != nil {
					return fmt.Errorf("can't POST to URL, err: %v", err)
				}
			}

			if metric.MType == "counter" {
				err := a.PostMetricURL(metric.MType, metric.ID, strconv.FormatInt(*metric.Delta, 10))
				if err != nil {
					return fmt.Errorf("can't POST to URL, err: %v", err)
				}
			}
		}

		done <- true
	}

	return nil
}

func (a *Agent) RunAgent(pollInterval, reportInterval int) {
	//rI := time.NewTicker(time.Duration(reportInterval) * time.Second)
	//pI := time.NewTicker(time.Duration(pollInterval) * time.Second)
	attempts := 3
	duration := 1

	var wg sync.WaitGroup

	g, _ := errgroup.WithContext(context.Background())

	go func(d time.Duration) {
		time.Sleep(d)
		a.Metrics.RuntimeMetricGenerator(runtime.MemStats{})
	}(time.Duration(a.Config.PollInterval) * time.Second)

	go func(d time.Duration) {
		time.Sleep(d)
		a.Metrics.GopsMetricGenerator()
	}(time.Duration(a.Config.PollInterval) * time.Second)

	go func(d time.Duration) {
		time.Sleep(d)
		a.Metrics.AdditionalMetricGenerator()
	}(time.Duration(a.Config.PollInterval) * time.Second)

	for {
		<-time.After(time.Duration(reportInterval) * time.Second)

		ms, errJSON := json.MetricListCreator(a.Metrics.Gauge, a.Metrics.Counter)

		if errJSON != nil {
			log.Print(errJSON)
		}

		jobs := make(chan json.Metrics, len(ms))
		done := make(chan bool, len(ms))

		for w := 0; w <= a.Config.RateLimit; w++ {
			wg.Add(1)
			g.Go(func() error {
				err := db.Retry(attempts, time.Duration(duration), func() error {
					err := a.PostMetrics("batch", jobs, done)
					return err
				})
				if err != nil {
					return err
				}
				return nil
			})
		}

		for _, m := range ms {
			jobs <- m
		}

		close(jobs)

		for range ms {
			<-done
		}
		wg.Done()
	}
}
