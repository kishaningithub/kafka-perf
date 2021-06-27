package cmd

import (
	"context"
	"fmt"
	"github.com/kishaningithub/kafka-perf/kafkaperf"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"os"
	"time"
)

var (
	reportType           string
	reportTimeStampField string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Outputs a report in the given format",
	RunE: func(cmd *cobra.Command, args []string) error {
		reportConfig := kafkaperf.ReporterConfig{
			Type:           reportType,
			TimeStampField: reportTimeStampField,
		}
		encoder := kafkaperf.NewEncoder(os.Stdin, kafkaperf.EncoderConfig{
			TimeStampField: reportTimeStampField,
		})
		metricsCalculator := kafkaperf.NewMetricsCalculator()
		reporter := kafkaperf.NewReporter(reportConfig, encoder, metricsCalculator)
		parentCtx, cancel := context.WithCancel(context.Background())
		operation, ctx := errgroup.WithContext(parentCtx)
		operation.Go(func() error {
			defer cancel()
			return reporter.GenerateReport(os.Stdout)
		})
		operation.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					stats := reporter.Stats()
					_, _ = fmt.Fprintf(os.Stderr, "\r%d records processed", stats.RecordsProcessed)
					return nil
				default:
					stats := reporter.Stats()
					_, _ = fmt.Fprintf(os.Stderr, "\r%d records processed", stats.RecordsProcessed)
				}
				time.Sleep(2 * time.Second)
			}
		})
		return operation.Wait()
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVar(&reportType, "type", "text", "Report type. Valid values are text")
	reportCmd.Flags().StringVar(&reportTimeStampField, "timestamp-field", "", "Field which has the unix timestamp. Eg 1617104831727")
	err := reportCmd.MarkFlagRequired("timestamp-field")
	if err != nil {
		panic(err)
	}
}
