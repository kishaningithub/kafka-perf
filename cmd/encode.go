package cmd

import (
	"github.com/kishaningithub/kafka-perf/kafkaperf"
	"github.com/pkg/profile"
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
		if Profile {
			defer profile.Start(profile.ProfilePath(".")).Stop()
		}
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
