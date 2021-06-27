package kafkaperf

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"log"
	"sync/atomic"
	"time"
)

type MonitorConfig struct {
	Topic            string
	BootstrapServers []string
	TlsMode          string
	CertLocation     string
	KeyLocation      string
	CACertLocation   string
}

const (
	TLS_MODE_NONE = "NONE"
	TLS_MODE_TLS  = "TLS"
	TLS_MODE_MTLS = "MTLS"
)

type Monitor interface {
	Start() error
}

type monitor struct {
	destination io.Writer
	kafkaReader *kafka.Reader
}

func NewMonitor(destination io.Writer, appConfig MonitorConfig) (Monitor, error) {
	baseErrMsg := "error while creating monitor"
	kafkaReaderConfig := kafka.ReaderConfig{
		Brokers:       appConfig.BootstrapServers,
		GroupID:       uuid.New().String(),
		Topic:         appConfig.Topic,
		StartOffset:   kafka.LastOffset,
		MinBytes:      1e6,  // 1MB
		MaxBytes:      10e6, // 10MB
		QueueCapacity: 10_000,
	}
	if appConfig.TlsMode != TLS_MODE_NONE {
		tlsConfig, err := getTLSConfig(appConfig)
		if err != nil {
			return nil, errors.Wrap(err, baseErrMsg)
		}
		kafkaReaderConfig.Dialer = &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS:       tlsConfig,
		}
	}
	kafkaReader := kafka.NewReader(kafkaReaderConfig)
	return &monitor{
		destination: destination,
		kafkaReader: kafkaReader,
	}, nil
}

type MonitorMetrics struct {
	NoOfEventsReceived  int64
	NoOfEventsRead      int64
	Lag                 int64
	NoOfEventsInProcess int64
	NoOfEventsProcessed int64
}

func (monitor *monitor) Start() (err error) {
	baseErrMsg := "error while monitoring kafka events"
	destination := monitor.destination
	var monitorMetrics MonitorMetrics
	events := make(chan string, DefaultChannelBufferSize)
	operation, _ := errgroup.WithContext(context.Background())
	operation.Go(func() error {
		for event := range events {
			atomic.AddInt64(&monitorMetrics.NoOfEventsInProcess, -1)
			_, err := fmt.Fprintln(destination, event)
			if err != nil {
				return errors.Wrap(err, baseErrMsg)
			}
			atomic.AddInt64(&monitorMetrics.NoOfEventsProcessed, 1)
		}
		return nil
	})
	operation.Go(func() error {
		defer close(events)
		for {
			message, err := monitor.kafkaReader.FetchMessage(context.Background())
			if err != nil {
				return errors.Wrap(err, baseErrMsg)
			}
			atomic.AddInt64(&monitorMetrics.NoOfEventsRead, 1)
			atomic.AddInt64(&monitorMetrics.NoOfEventsInProcess, 1)
			marshal, err := json.Marshal(message)
			if err != nil {
				return errors.Wrap(err, baseErrMsg)
			}
			events <- string(marshal)
		}
	})
	operation.Go(func() error {
		for {
			time.Sleep(2 * time.Second)
			stats := monitor.kafkaReader.Stats()
			monitorMetrics.Lag = stats.Lag
			atomic.AddInt64(&monitorMetrics.NoOfEventsReceived, stats.Messages)
			jsonMetrics, _ := json.Marshal(monitorMetrics)
			log.Println(string(jsonMetrics))
		}
	})
	return operation.Wait()
}

func getTLSConfig(config MonitorConfig) (*tls.Config, error) {
	caCertLocation := config.CACertLocation
	caCert, err := ioutil.ReadFile(caCertLocation)
	if err != nil {
		return nil, errors.Wrapf(err, "error while loading ca cert from %s", caCertLocation)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS12,
	}
	if config.TlsMode == TLS_MODE_MTLS {
		certLocation := config.CertLocation
		keyLocation := config.KeyLocation
		cert, err := tls.LoadX509KeyPair(certLocation, keyLocation)
		if err != nil {
			return nil, errors.Wrapf(err, "error while loading cert %s and key %s", certLocation, keyLocation)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return tlsConfig, nil
}
