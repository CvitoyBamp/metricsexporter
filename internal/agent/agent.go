package agent

import (
	"bytes"
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/crypto"
	"github.com/CvitoyBamp/metricsexporter/internal/db"
	"github.com/CvitoyBamp/metricsexporter/internal/json"
	"github.com/CvitoyBamp/metricsexporter/internal/metrics"
	"github.com/CvitoyBamp/metricsexporter/internal/middlewares"
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"log"
	"net/http"
	"runtime"
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
	Metrics  *storage.MemStorage
	Config   *Config
}

func CreateAgent(cfg Config) *Agent {
	return &Agent{
		Client: &http.Client{
			Timeout: 1 * time.Second,
		},
		Endpoint: cfg.Address,
		Metrics:  storage.CreateMemStorage(),
		Config:   &cfg,
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
		req.Header.Set("HashSHA256", crypto.CreateHash(compressedData, a.Config.Key))
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

func (a *Agent) PostMetricsBatch(ch chan storage.MemStorage) error {
	m := <-ch
	data, errJSON := json.ListCreator(m.Gauge, m.Counter)
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
		req.Header.Set("HashSHA256", crypto.CreateHash(compressedData, a.Config.Key))
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

//func (a *Agent) PostMetrics(types string, i json.Metrics) error {
//	a.Metrics.RLock()
//	defer a.Metrics.RUnlock()
//
//	if types == "batch" {
//		err := a.PostMetricsBatch()
//		if err != nil {
//			return fmt.Errorf("can't POST to URL, err: %v", err)
//		}
//	}
//
//		if types == "json" {
//			if i.MType == "gauge" {
//				err := a.PostMetricJSON(i.MType, i.ID, strconv.FormatFloat(*i.Value, 'f', -1, 64))
//				if err != nil {
//					return fmt.Errorf("can't POST to URL, err: %v", err)
//				}
//			}
//			if i.MType == "counter" {
//				err := a.PostMetricJSON(i.MType, i.ID, strconv.FormatInt(*i.Delta, 10))
//				if err != nil {
//					return fmt.Errorf("can't POST to URL, err: %v", err)
//				}
//			}
//		}
//		if types == "url" {
//			if i.MType == "gauge" {
//				err := a.PostMetricURL(i.MType, i.ID, strconv.FormatFloat(*i.Value, 'f', -1, 64))
//				if err != nil {
//					return fmt.Errorf("can't POST to URL, err: %v", err)
//				}
//			}
//			if i.MType == "counter" {
//				err := a.PostMetricURL(i.MType, i.ID, strconv.FormatInt(*i.Delta, 10))
//				if err != nil {
//					return fmt.Errorf("can't POST to URL, err: %v", err)
//				}
//			}
//		}
//	return nil
//}

func (a *Agent) worker(types string, job chan storage.MemStorage) error {

	if types == "batch" {
		err := a.PostMetricsBatch(job)
		if err != nil {
			return fmt.Errorf("can't POST to URL, err: %v", err)
		}
	}

	//for i := range job {
	//	if types == "json" {
	//		if i.MType == "gauge" {
	//			err := a.PostMetricJSON(i.MType, i.ID, strconv.FormatFloat(*i.Value, 'f', -1, 64))
	//			if err != nil {
	//				return fmt.Errorf("can't POST to URL, err: %v", err)
	//			}
	//		}
	//		if i.MType == "counter" {
	//			err := a.PostMetricJSON(i.MType, i.ID, strconv.FormatInt(*i.Delta, 10))
	//			if err != nil {
	//				return fmt.Errorf("can't POST to URL, err: %v", err)
	//			}
	//		}
	//	}
	//	if types == "url" {
	//		if i.MType == "gauge" {
	//			err := a.PostMetricURL(i.MType, i.ID, strconv.FormatFloat(*i.Value, 'f', -1, 64))
	//			if err != nil {
	//				return fmt.Errorf("can't POST to URL, err: %v", err)
	//			}
	//		}
	//		if i.MType == "counter" {
	//			err := a.PostMetricURL(i.MType, i.ID, strconv.FormatInt(*i.Delta, 10))
	//			if err != nil {
	//				return fmt.Errorf("can't POST to URL, err: %v", err)
	//			}
	//		}
	//	}
	//}
	return nil
}

func (a *Agent) RunAgent() {
	rI := time.NewTicker(time.Duration(a.Config.ReportInterval) * time.Second)
	//pI := time.NewTicker(time.Duration(a.Config.PollInterval) * time.Second)
	attempts := 3
	duration := 1

	job := make(chan storage.MemStorage, a.Config.RateLimit)
	defer close(job)

	wg := sync.WaitGroup{}

	wg.Add(1)
	func(duration time.Duration) {
		for {
			time.Sleep(duration)
			go metrics.MetricGenerator(runtime.MemStats{}, a.Metrics, job)
			log.Println(<-job)
		}
	}(time.Duration(a.Config.PollInterval) * time.Second)

	for {
		select {
		case <-rI.C:
			for w := 1; w <= a.Config.RateLimit; w++ {
				err := db.Retry(attempts, time.Duration(duration), func() error {
					err := a.worker("batch", job)
					return err
				})
				if err != nil {
					log.Println(err)
				}

			}
		}
	}

	wg.Wait()
}
