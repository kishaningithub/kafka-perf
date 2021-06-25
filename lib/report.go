package monitor

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/linkedin/goavro"
	"github.com/montanaflynn/stats"
	"github.com/segmentio/kafka-go"
	"io"
	"strconv"
	"time"
)

type ReportConfig struct {
	Type           string
	TimeStampField string
}

type Reporter interface {
	GenerateReport(writer io.Writer) error
}

type reporter struct {
	source       io.Reader
	reportConfig ReportConfig
}

func NewReporter(source io.Reader, reportConfig ReportConfig) Reporter {
	return &reporter{
		source:       source,
		reportConfig: reportConfig,
	}
}

func (reporter *reporter) GenerateReport(writer io.Writer) error {
	data, err := reporter.aggregateData()
	if err != nil {
		return err
	}
	report := reporter.textReport(data)
	_, err = io.WriteString(writer, report)
	if err != nil {
		return err
	}
	return nil
}

type AggregateData struct {
	totalNoOfEvents       int
	mean                  float64
	percentile95          float64
	percentile99          float64
	partitionDistribution map[int]int
}

func (reporter *reporter) textReport(aggregateData AggregateData) string {
	var report string
	report += fmt.Sprintf(`
Total No. Of events %d
Latency
=======
Mean                %f
95th Percentile     %f
99th Percentile     %f
`, aggregateData.totalNoOfEvents, aggregateData.mean, aggregateData.percentile95, aggregateData.percentile99)

	report += fmt.Sprintf(`
Partition
=========
`)

	for partitionKey, noOfMessages := range aggregateData.partitionDistribution {
		report += fmt.Sprintf("partition %d  %d messages\n", partitionKey, noOfMessages)
	}

	return report
}

func (reporter *reporter) aggregateData() (AggregateData, error) {
	scanner := bufio.NewScanner(reporter.source)
	latencies := make(stats.Float64Data, 0, 1000)
	partitionDistribution := make(map[int]int)
	for scanner.Scan() {
		text := scanner.Text()
		var kafkaMessage kafka.Message
		err := json.Unmarshal([]byte(text), &kafkaMessage)
		if err != nil {
			return AggregateData{}, err
		}
		partitionDistribution[kafkaMessage.Partition]++
		ocfReader, err := goavro.NewOCFReader(bytes.NewBuffer(kafkaMessage.Value))
		if err != nil {
			return AggregateData{}, err
		}
		for ocfReader.Scan() {
			record, err := ocfReader.Read()
			if err != nil {
				return AggregateData{}, err
			}
			messageSentTime, err := reporter.extractTimeStampFromAvroData(record, err)
			if err != nil {
				return AggregateData{}, err
			}
			messageReceivedTime := kafkaMessage.Time
			latency := float64(messageReceivedTime.Sub(messageSentTime)) / float64(time.Millisecond)
			latencies = append(latencies, latency)
		}
	}

	mean, err := stats.Mean(latencies)
	if err != nil {
		return AggregateData{}, err
	}
	percentile95, err := stats.Percentile(latencies, 95)
	if err != nil {
		return AggregateData{}, err
	}
	percentile99, err := stats.Percentile(latencies, 99)
	if err != nil {
		return AggregateData{}, err
	}
	totalNoOfEvents := len(latencies)
	return AggregateData{
		totalNoOfEvents:       totalNoOfEvents,
		mean:                  mean,
		percentile95:          percentile95,
		percentile99:          percentile99,
		partitionDistribution: partitionDistribution,
	}, nil
}

func (reporter *reporter) extractTimeStampFromAvroData(record interface{}, err error) (time.Time, error) {
	recordMap, ok := record.(map[string]interface{})
	if !ok {
		return time.Time{}, fmt.Errorf("invalid data %v", record)
	}
	epochMillisecondStr := fmt.Sprintf("%v", recordMap[reporter.reportConfig.TimeStampField])
	epochMillisecond, err := strconv.Atoi(epochMillisecondStr)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(epochMillisecond), 0), nil
}
