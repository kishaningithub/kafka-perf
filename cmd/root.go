package cmd

import (
	"github.com/spf13/cobra"
)

var Version string
var Verbose bool
var Profile bool

var rootCmd = &cobra.Command{
	Use:   "kafka-perf",
	Short: "Get performance metrics based on kafka events",
	Long: `Supported metrics
1. Latency - Mean, 95th percentile, 99th percentile
2. Total No of events
3. Partition distribution
`,
	Version: Version,
}

func init() {
	rootCmd.AddCommand(reportCmd)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&Profile, "profile", false, "profile this cli tool")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
