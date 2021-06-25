package cmd

import (
	"github.com/kishaningithub/kafka-perf/kafkaperf"
	"github.com/spf13/cobra"
	"os"
)

var (
	reportType     string
	timeStampField string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Outputs a report in the given format",
	RunE: func(cmd *cobra.Command, args []string) error {
		reportConfig := kafkaperf.ReportConfig{
			Type:           reportType,
			TimeStampField: timeStampField,
		}
		reporter := kafkaperf.NewReporter(os.Stdin, reportConfig)
		return reporter.GenerateReport(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVar(&reportType, "type", "text", "Report type. Valid values are text")
	reportCmd.Flags().StringVar(&timeStampField, "timestamp-field", "", "Field which has the unix timestamp. Eg 1617104831727")
	err := reportCmd.MarkFlagRequired("timestamp-field")
	if err != nil {
		panic(err)
	}
}
