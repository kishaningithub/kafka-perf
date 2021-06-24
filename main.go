package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type MonitConfig struct {
	Topic            string
	BootstrapServers []string
	TlsMode          string
	CertLocation     string
	KeyLocation      string
	CACertLocation   string
}

type ReportConfig struct {
	Type string
}

const (
	TLS_MODE_NONE = "NONE"
	TLS_MODE_TLS  = "TLS"
	TLS_MODE_MTLS = "MTLS"
)

func main() {
	monitCmd := flag.NewFlagSet("monit", flag.ExitOnError)
	topic := monitCmd.String("topic", "", "REQUIRED: The topic id to consume on.")
	bootstrapServersCSV := monitCmd.String("bootstrap-servers", "", "REQUIRED: The server(s) to connect to.")
	tlsMode := monitCmd.String("tls-mode", "NONE", "Valid values are NONE,TLS,MTLS")
	certLocation := monitCmd.String("tls-cert", "", "certificate file location. Eg. /certs/cert.pem. Required if tls-mode is MTLS")
	keyLocation := monitCmd.String("tls-key", "", "key file location. Eg. /certs/key.pem. Required if tls-mode is MTLS")
	caCertLocation := monitCmd.String("tls-ca-cert", "", "CA cert file location. Eg. /certs/ca.pem. Required if tls-mode is TLS, MTLS")

	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	reportType := reportCmd.String("type", "text", "Report type. Valid values are text")

	if len(os.Args) < 2 {
		_, _ = os.Stderr.WriteString("expected 'monit' or 'report' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "monit":
		err := monitCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
		monitConfig := MonitConfig{
			Topic:            *topic,
			BootstrapServers: strings.Split(*bootstrapServersCSV, ","),
			TlsMode:          *tlsMode,
			CertLocation:     *certLocation,
			KeyLocation:      *keyLocation,
			CACertLocation:   *caCertLocation,
		}
		_, _ = os.Stderr.WriteString(fmt.Sprintf("loaded config %+v \n", monitConfig))
		monitor(monitConfig)
	case "report":
		err := reportCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
		reportConfig := ReportConfig{
			Type: *reportType,
		}
		_, _ = os.Stderr.WriteString(fmt.Sprintf("loaded config %+v \n", reportConfig))
		report(reportConfig)
	default:
		_, _ = os.Stderr.WriteString("expected 'monit' or 'report' subcommands")
		os.Exit(1)
	}
}

func getTLSConfig(config MonitConfig) *tls.Config {
	caCertLocation := config.CACertLocation
	caCert, err := ioutil.ReadFile(caCertLocation)
	if err != nil {
		panic(fmt.Errorf("error while loading ca cert from %s: %w", caCertLocation, err))
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
			panic(fmt.Errorf("error while loading cert %s and key %s: %w", certLocation, keyLocation, err))
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	return tlsConfig
}
