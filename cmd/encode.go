package cmd

import (
	"github.com/kishaningithub/kafka-perf/kafkaperf"
	"os"

	"github.com/spf13/cobra"
)

var (
	encodeType           string
	encodeTimeStampField string
)

var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Encode results from one encoding to another",
	RunE: func(cmd *cobra.Command, args []string) error {
		encoder := kafkaperf.NewEncoder(os.Stdin, kafkaperf.EncoderConfig{
			Type:           encodeType,
			TimeStampField: encodeTimeStampField,
		})
		return encoder.Encode(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(encodeCmd)

	encodeCmd.Flags().StringVar(&encodeType, "type", "csv", "Report type. Valid values are csv")
	encodeCmd.Flags().StringVar(&encodeTimeStampField, "timestamp-field", "", "Field which has the unix timestamp. Eg 1617104831727")
	err := encodeCmd.MarkFlagRequired("timestamp-field")
	if err != nil {
		panic(err)
	}
}
