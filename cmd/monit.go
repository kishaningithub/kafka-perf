package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kishaningithub/kafka-perf/kafkaperf"
	"github.com/pkg/profile"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"strings"
	"time"

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
		if Profile {
			defer profile.Start(profile.ProfilePath(".")).Stop()
		}
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
		operation, _ := errgroup.WithContext(context.Background())
		operation.Go(func() error {
			return monit.Start()
		})
		operation.Go(func() error {
			for {
				stats := monit.Stats()
				if Verbose {
					statsJson, _ := json.Marshal(stats)
					log.Println(string(statsJson))
				} else {
					_, _ = fmt.Fprintf(os.Stderr, "\r%d events processed", stats.EventsProcessed)
				}
				time.Sleep(2 * time.Second)
			}
		})
		return operation.Wait()
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
