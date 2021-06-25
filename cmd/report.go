package cmd

import (
	"github.com/kishaningithub/kafka-perf/kafkaperf"
	"github.com/spf13/cobra"
	"os"
)

var (
	reportType           string
	reportTimeStampField string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Outputs a report in the given format",
	RunE: func(cmd *cobra.Command, args []string) error {
		reportConfig := kafkaperf.ReportConfig{
			Type:           reportType,
			TimeStampField: reportTimeStampField,
		}
		encoder := kafkaperf.NewEncoder(os.Stdin, kafkaperf.EncoderConfig{
			TimeStampField: reportTimeStampField,
		})
		metricsCalculator := kafkaperf.NewMetricsCalculator()
		reporter := kafkaperf.NewReporter(reportConfig, encoder, metricsCalculator)
		return reporter.GenerateReport(os.Stdout)
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
