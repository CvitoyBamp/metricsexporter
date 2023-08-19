package handlers

import (
	"bufio"
	"github.com/CvitoyBamp/metricsexporter/internal/db"
	"github.com/CvitoyBamp/metricsexporter/internal/json"
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"log"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	DSN           string `env:"DATABASE_DSN"`
	Key           string `env:"KEY"`
}

type Consumer struct {
	file   *os.File
	reader *bufio.Reader
}

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

type CustomServer struct {
	Server  *http.Server
	Storage *storage.MemStorage
	Config  *Config
	DB      *db.Database
}

func CreateServer(cfg Config) *CustomServer {
	return &CustomServer{
		Server:  &http.Server{},
		Storage: storage.CreateMemStorage(),
		Config:  &cfg,
		DB:      db.CreateDB(cfg.DSN),
	}
}

func (s *CustomServer) PreloadMetrics() error {
	consumer, err := s.newConsumer()
	if err != nil {
		return err
	}
	errRead := consumer.readFromFile(s.Storage)
	if errRead != nil {
		return errRead
	}
	return nil
}

func (s *CustomServer) RunServer() error {
	return http.ListenAndServe(s.Config.Address, s.MetricRouter())
}

func (s *CustomServer) PostSaveMetrics() {

	sI := time.NewTicker(time.Duration(s.Config.StoreInterval) * time.Second)

	for {
		<-sI.C
		producer, errProducer := s.newProducer(false)
		if errProducer != nil {
			log.Print(errProducer)
		}
		errSave := producer.saveToFile(s.Storage)
		if errProducer != nil {
			log.Print(errSave)
		}
	}
}

func (s *CustomServer) StopServer() error {
	return s.Server.Close()
}

func (s *CustomServer) newProducer(sync bool) (*Producer, error) {

	var file *os.File
	var err error

	if sync {
		file, err = os.OpenFile(s.Config.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 0666)
		if err != nil {
			return nil, err
		}
	} else {
		file, err = os.OpenFile(s.Config.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, err
		}
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (s *CustomServer) newConsumer() (*Consumer, error) {
	file, err := os.OpenFile(s.Config.FilePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (p *Producer) saveToFile(ms *storage.MemStorage) error {
	data, errData := json.MetricConverter(ms)

	if errData != nil {
		return errData
	}

	if _, errWrite := p.writer.Write(data); errWrite != nil {
		return errWrite
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	log.Print("put the metrics in a file")

	return p.writer.Flush()
}

func (c *Consumer) readFromFile(ms *storage.MemStorage) error {

	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	errJSON := json.Decoder(data, ms)
	if errJSON != nil {
		return errJSON
	}

	log.Print("i've read metrics from file")

	return nil
}
