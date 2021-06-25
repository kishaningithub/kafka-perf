package cmd

import (
	"github.com/kishaningithub/kafka-perf/kafkaperf"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	topic               string
	bootstrapServersCSV string
	tlsMode             string
	certLocation        string
	keyLocation         string
	caCertLocation      string
)

var monitCmd = &cobra.Command{
	Use:   "monit",
	Short: "Monitoring kafka events",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := kafkaperf.MonitorConfig{
			Topic:            topic,
			BootstrapServers: strings.Split(bootstrapServersCSV, ","),
			TlsMode:          tlsMode,
			CertLocation:     certLocation,
			KeyLocation:      keyLocation,
			CACertLocation:   caCertLocation,
		}
		monit, err := kafkaperf.NewMonitor(os.Stdout, config)
		if err != nil {
			return err
		}
		return monit.Start()
	},
}

func init() {
	rootCmd.AddCommand(monitCmd)

	monitCmd.Flags().StringVar(&topic, "topic", "", "REQUIRED: The topic id to consume on")
	err := monitCmd.MarkFlagRequired("topic")
	if err != nil {
		panic(err)
	}
	monitCmd.Flags().StringVar(&bootstrapServersCSV, "bootstrap-servers", "", "REQUIRED: The server(s) to connect to.")
	err = monitCmd.MarkFlagRequired("bootstrap-servers")
	if err != nil {
		panic(err)
	}
	monitCmd.Flags().StringVar(&tlsMode, "tls-mode", "NONE", "Valid values are NONE,TLS,MTLS")
	monitCmd.Flags().StringVar(&certLocation, "tls-cert", "", "certificate file location. Eg. /certs/cert.pem. Required if tls-mode is MTLS")
	monitCmd.Flags().StringVar(&keyLocation, "tls-key", "", "key file location. Eg. /certs/key.pem. Required if tls-mode is MTLS")
	monitCmd.Flags().StringVar(&caCertLocation, "tls-ca-cert", "", "CA cert file location. Eg. /certs/ca.pem. Required if tls-mode is TLS, MTLS")
}
