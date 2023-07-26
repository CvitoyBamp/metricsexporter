package handlers

import (
	"bufio"
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"github.com/CvitoyBamp/metricsexporter/internal/util"
	"log"
	"net/http"
	"os"
	"time"
)

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
}

func CreateServer() *CustomServer {
	return &CustomServer{
		Server:  &http.Server{},
		Storage: storage.CreateMemStorage(),
	}
}

func (s *CustomServer) PreloadMetrics(filename string) error {
	consumer, err := s.newConsumer(filename)
	if err != nil {
		return err
	}
	errRead := consumer.readFromFile(s.Storage)
	if errRead != nil {
		return errRead
	}
	return nil
}

func (s *CustomServer) RunServer(address string) error {
	return http.ListenAndServe(address, s.MetricRouter())
}

func (s *CustomServer) PostSaveMetrics(filename string, storeInterval int) error {

	if storeInterval == 0 {
		producer, err := s.newProducer(filename, storeInterval)
		log.Print(err)
		log.Print(producer.saveToFile(s.Storage))
	} else {
		sI := time.NewTicker(time.Duration(storeInterval) * time.Second)

		for {
			select {
			case <-sI.C:
				producer, err := s.newProducer(filename, storeInterval)
				log.Print(err)
				log.Print(producer.saveToFile(s.Storage))
			}
		}
	}

	return nil
}

func (s *CustomServer) StopServer() error {
	return s.Server.Close()
}

func (s *CustomServer) newProducer(filename string, storeInterval int) (*Producer, error) {

	var file *os.File
	var err error

	if storeInterval == 0 {
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 0666)
		if err != nil {
			return nil, err
		}
	} else {
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, err
		}
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (s *CustomServer) newConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (p *Producer) saveToFile(ms *storage.MemStorage) error {
	data, errData := util.JSONMetricConverter(ms)

	if errData != nil {
		return errData
	}

	if _, errWrite := p.writer.Write(data); errWrite != nil {
		return errWrite
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

func (c *Consumer) readFromFile(ms *storage.MemStorage) error {

	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	errJSON := util.JSONDecoder(data, ms)
	if errJSON != nil {
		return errJSON
	}

	log.Print("i've read metrics from file")

	return nil
}
