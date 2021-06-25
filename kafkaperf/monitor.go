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
	"io"
	"io/ioutil"
	"os"
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
		Brokers:     appConfig.BootstrapServers,
		GroupID:     uuid.New().String(),
		Topic:       appConfig.Topic,
		StartOffset: kafka.LastOffset,
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

func (monitor *monitor) Start() error {
	baseErrMsg := "error while monitoring kafka events"
	for i := 0; ; i++ {
		message, err := monitor.kafkaReader.ReadMessage(context.Background())
		if err != nil {
			return errors.Wrap(err, baseErrMsg)
		}
		marshal, err := json.Marshal(message)
		if err != nil {
			return errors.Wrap(err, baseErrMsg)
		}
		_, err = fmt.Fprintln(monitor.destination, string(marshal))
		if err != nil {
			return errors.Wrap(err, baseErrMsg)
		}
		_, _ = fmt.Fprintf(os.Stderr, "\r%d events stored", i)
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
