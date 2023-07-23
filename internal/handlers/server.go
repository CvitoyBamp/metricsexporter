package handlers

import (
	"bufio"
	"fmt"
	"github.com/CvitoyBamp/metricsexporter/internal/storage"
	"github.com/CvitoyBamp/metricsexporter/internal/util"
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

	sI := time.NewTicker(time.Duration(storeInterval) * time.Second)

	producer, err := s.newProducer(filename)
	if err != nil {
		return err
	}

	for {
		select {
		case <-sI.C:
			errSave := producer.saveToFile(s.Storage)
			if errSave != nil {
				break
			}
		}

	}

	return fmt.Errorf("can't save to file")
}

func (s *CustomServer) StopServer() error {
	return s.Server.Close()
}

func (s *CustomServer) newProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
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

	errJson := util.JSONDecoder(data, ms)
	if errJson != nil {
		return errJson
	}

	return nil
}
