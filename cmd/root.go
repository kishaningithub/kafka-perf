package cmd

import (
	"github.com/spf13/cobra"
)

var Version string

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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
