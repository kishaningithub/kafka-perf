package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/linkedin/goavro"
	"github.com/montanaflynn/stats"
)

func report(appConfig ReportConfig) {
	ocfReader, err := goavro.NewOCFReader(bytes.NewBuffer(nil))
	if err != nil {
		panic(fmt.Errorf("data is not in plain avro format %s: %w", "", err))
	}
	avroSchema := ocfReader.Codec().Schema()
	fmt.Println("Schema")
	fmt.Println("=====")
	fmt.Println(avroSchema)
	fmt.Println("Data")
	fmt.Println("=====")
	for ocfReader.Scan() {
		record, _ := ocfReader.Read()
		jsonRecord, err := json.Marshal(record)
		if err != nil {
			panic(fmt.Errorf("unable to marshal record %v as json: %w", record, err))
		}
		fmt.Println(string(jsonRecord))
	}
	fmt.Println()

	latencies := []float64{43, 54, 56, 61, 62, 66}
	percentile90, _ := stats.Percentile(latencies, 95)
	percentile99, _ := stats.Percentile(latencies, 99)
	mean, _ := stats.Mean(latencies)
	fmt.Println(percentile90, percentile99, mean)
}
