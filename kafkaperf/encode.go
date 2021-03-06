package kafkaperf

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/gocarina/gocsv"
	"github.com/linkedin/goavro"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"io"
	"runtime"
	"strconv"
	"time"
)

type EncoderConfig struct {
	Type           string
	TimeStampField string
}

type RawMetricsData struct {
	MessageSentTime     time.Time
	MessageReceivedTime time.Time
	Partition           int
}

func (rawMetricsData RawMetricsData) Latency() time.Duration {
	return rawMetricsData.MessageReceivedTime.Sub(rawMetricsData.MessageSentTime)
}

type CSVEncodedData struct {
	MessageSentTime      time.Time     `csv:"MessageSentTime"`
	MessageReceivedTime  time.Time     `csv:"MessageReceivedTime"`
	Latency              time.Duration `csv:"Latency"`
	LatencyInNanoSeconds int           `csv:"LatencyInNanoSeconds"`
	Partition            int           `csv:"Partition"`
}

func NewCSVEncodedData(rawMetricsData RawMetricsData) CSVEncodedData {
	latency := rawMetricsData.Latency()
	return CSVEncodedData{
		MessageSentTime:      rawMetricsData.MessageSentTime,
		MessageReceivedTime:  rawMetricsData.MessageReceivedTime,
		Latency:              latency,
		LatencyInNanoSeconds: int(latency.Nanoseconds()),
		Partition:            rawMetricsData.Partition,
	}
}

type Encoder interface {
	EncodeAsStruct(result chan<- RawMetricsData) error
	Encode(destination io.Writer) error
}

type encoder struct {
	source        io.Reader
	encoderConfig EncoderConfig
}

func NewEncoder(source io.Reader, encoderConfig EncoderConfig) Encoder {
	return &encoder{
		source:        source,
		encoderConfig: encoderConfig,
	}
}

func (encoder *encoder) EncodeAsStruct(result chan<- RawMetricsData) error {
	defer close(result)
	baseErrMsg := "error while extracting raw metrics data"
	lines := make(chan string, DefaultChannelBufferSize)
	operation, _ := errgroup.WithContext(context.Background())

	operation.Go(func() error {
		defer close(lines)
		baseErrMsg := "error while reading record"
		scanner := bufio.NewScanner(bufio.NewReader(encoder.source))
		maxLineLengthInBytes := 10 * 1024 * 1024
		buf := make([]byte, maxLineLengthInBytes)
		scanner.Buffer(buf, maxLineLengthInBytes)
		for scanner.Scan() {
			line := scanner.Text()
			lines <- line
		}
		if err := scanner.Err(); err != nil {
			return errors.Wrap(err, baseErrMsg)
		}
		return nil
	})
	for i := 0; i < runtime.NumCPU(); i++ {
		operation.Go(func() error {
			for line := range lines {
				err := encoder.processLine(line, result)
				if err != nil {
					return errors.Wrap(err, baseErrMsg)
				}
			}
			return nil
		})
	}
	return operation.Wait()
}

func (encoder *encoder) processLine(line string, result chan<- RawMetricsData) error {
	var kafkaMessage kafka.Message
	err := json.Unmarshal([]byte(line), &kafkaMessage)
	if err != nil {
		return err
	}
	ocfReader, err := goavro.NewOCFReader(bytes.NewBuffer(kafkaMessage.Value))
	if err != nil {
		return err
	}
	for ocfReader.Scan() {
		record, err := ocfReader.Read()
		if err != nil {
			return err
		}
		messageSentTime, err := encoder.extractTimeStampFromAvroData(record)
		if err != nil {
			return err
		}
		messageReceivedTime := kafkaMessage.Time
		rawMetricsData := RawMetricsData{
			MessageSentTime:     messageSentTime.UTC(),
			MessageReceivedTime: messageReceivedTime.UTC(),
			Partition:           kafkaMessage.Partition,
		}
		result <- rawMetricsData
	}
	return nil
}

func (encoder *encoder) Encode(destination io.Writer) error {
	csvWriter := make(chan interface{}, DefaultChannelBufferSize)
	operation, _ := errgroup.WithContext(context.Background())
	operation.Go(func() error {
		return gocsv.MarshalChan(csvWriter, gocsv.DefaultCSVWriter(destination))
	})
	operation.Go(func() error {
		defer close(csvWriter)
		metricsDataStream := make(chan RawMetricsData, DefaultChannelBufferSize)
		operation.Go(func() error {
			return encoder.EncodeAsStruct(metricsDataStream)
		})
		for metricsData := range metricsDataStream {
			csvWriter <- NewCSVEncodedData(metricsData)
		}
		return nil
	})
	return operation.Wait()
}

func (encoder *encoder) extractTimeStampFromAvroData(record interface{}) (time.Time, error) {
	baseErrMsg := fmt.Sprintf("error while extracting timestamp from avro data from record %v", record)
	recordMap, ok := record.(map[string]interface{})
	if !ok {
		return time.Time{}, errors.WithMessage(errors.New("incorrect data format. expecting a map[string]interface{}"), baseErrMsg)
	}
	epochMillisecondStr := fmt.Sprintf("%v", recordMap[encoder.encoderConfig.TimeStampField])
	epochMillisecond, err := strconv.Atoi(epochMillisecondStr)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "%s %s", baseErrMsg)
	}
	return time.Unix(0, 0).Add(time.Duration(epochMillisecond) * time.Millisecond), nil
}
