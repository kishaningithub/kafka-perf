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
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
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
	Stats() MonitorMetrics
}

type monitor struct {
	destination    io.Writer
	kafkaReader    *kafka.Reader
	monitorMetrics *monitorMetrics
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
		QueueCapacity: 10e3,
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
		destination:    destination,
		kafkaReader:    kafkaReader,
		monitorMetrics: &monitorMetrics{},
	}, nil
}

type MonitorMetrics struct {
	EventsReceived  int64
	EventsRead      int64
	Lag             int64
	EventsInProcess int64
	EventsProcessed int64
}

type monitorMetrics struct {
	eventsReceived  atomic.Int64
	eventsRead      atomic.Int64
	eventsInProcess atomic.Int64
	eventsProcessed atomic.Int64
}

func (monitor *monitor) Start() (err error) {
	baseErrMsg := "error while monitoring kafka events"
	destination := monitor.destination
	monitorMetrics := monitor.monitorMetrics
	events := make(chan string, DefaultChannelBufferSize)
	operation, _ := errgroup.WithContext(context.Background())
	operation.Go(func() error {
		for event := range events {
			monitorMetrics.eventsInProcess.Sub(1)
			_, err := fmt.Fprintln(destination, event)
			if err != nil {
				return errors.Wrap(err, baseErrMsg)
			}
			monitorMetrics.eventsProcessed.Add(1)
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
			monitorMetrics.eventsRead.Add(1)
			monitorMetrics.eventsInProcess.Add(1)
			marshal, err := json.Marshal(message)
			if err != nil {
				return errors.Wrap(err, baseErrMsg)
			}
			events <- string(marshal)
		}
	})
	return operation.Wait()
}

func (monitor *monitor) Stats() MonitorMetrics {
	stats := monitor.kafkaReader.Stats()
	monitorMetrics := monitor.monitorMetrics
	monitorMetrics.eventsReceived.Add(stats.Messages)
	return MonitorMetrics{
		EventsReceived:  monitorMetrics.eventsReceived.Load(),
		EventsRead:      monitorMetrics.eventsRead.Load(),
		Lag:             stats.Lag,
		EventsInProcess: monitorMetrics.eventsInProcess.Load(),
		EventsProcessed: monitorMetrics.eventsProcessed.Load(),
	}
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
