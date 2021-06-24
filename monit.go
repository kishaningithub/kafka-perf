package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"time"
)

func monitor(appConfig MonitConfig) {
	kafkaReaderConfig := kafka.ReaderConfig{
		Brokers: appConfig.BootstrapServers,
		GroupID: "monitGroupID",
		Topic:   appConfig.Topic,
	}
	if appConfig.TlsMode != TLS_MODE_NONE {
		tlsConfig := getTLSConfig(appConfig)
		kafkaReaderConfig.Dialer = &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS:       tlsConfig,
		}
	}
	reader := kafka.NewReader(kafkaReaderConfig)
	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			panic(fmt.Errorf("error while reading message from kafka topic %s: %w", appConfig.Topic, err))
		}
		marshal, err := json.Marshal(message)
		if err != nil {
			panic(fmt.Errorf("error while reading message from kafka topic %s: %w", appConfig.Topic, err))
		}
		fmt.Println(string(marshal))
	}
}
