package kafkaperf

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
)

type ReportConfig struct {
	Type           string
	TimeStampField string
}

type Reporter interface {
	GenerateReport(writer io.Writer) error
}

type reporter struct {
	reportConfig      ReportConfig
	encoder           Encoder
	metricsCalculator MetricsCalculator
}

func NewReporter(reportConfig ReportConfig, encoder Encoder, metricsCalculator MetricsCalculator) Reporter {
	return &reporter{
		reportConfig:      reportConfig,
		encoder:           encoder,
		metricsCalculator: metricsCalculator,
	}
}

func (reporter *reporter) GenerateReport(writer io.Writer) error {
	baseErrMsg := "error while generating report"
	data, err := reporter.aggregateData()
	if err != nil {
		return errors.Wrap(err, baseErrMsg)
	}
	report := reporter.textReport(data)
	_, err = io.WriteString(writer, report)
	if err != nil {
		return errors.Wrap(err, baseErrMsg)
	}
	return nil
}

func (reporter *reporter) textReport(aggregateData Metrics) string {
	baseFormat := `
=======
Report
=======

Total No. of events %v

Latency
=======
Mean                %v
95th Percentile     %v
99th Percentile     %v

Partition
=========
`
	partitionFormat := `
partition %v        %v messages
`
	var report string
	report += fmt.Sprintf(baseFormat, aggregateData.totalNoOfEvents, aggregateData.latencyMetrics.mean, aggregateData.latencyMetrics.percentile95, aggregateData.latencyMetrics.percentile99)
	for partitionKey, noOfMessages := range aggregateData.partitionDistribution {
		report += fmt.Sprintf(partitionFormat, partitionKey, noOfMessages)
	}
	return report
}

func (reporter *reporter) aggregateData() (Metrics, error) {
	rawMetricsDataStream := make(chan RawMetricsData, DefaultChannelBufferSize)
	operation, _ := errgroup.WithContext(context.Background())
	operation.Go(func() error {
		return reporter.encoder.EncodeAsStruct(rawMetricsDataStream)
	})
	i := 0
	for rawMetricsData := range rawMetricsDataStream {
		reporter.metricsCalculator.AddRawMetric(rawMetricsData)
		i++
		_, _ = fmt.Fprintf(os.Stderr, "\r%d events processed", i)
	}
	return reporter.metricsCalculator.GetMetrics(), operation.Wait()
}
