package kafkaperf

import (
	"github.com/influxdata/tdigest"
	"time"
)

type Metrics struct {
	totalNoOfEvents       int
	latencyMetrics        LatencyMetrics
	partitionDistribution map[int]int
}

type LatencyMetrics struct {
	mean         time.Duration
	percentile95 time.Duration
	percentile99 time.Duration
	total        time.Duration
}

type MetricsCalculator interface {
	AddRawMetric(rawMetricsData RawMetricsData)
	GetMetrics() Metrics
}

type metricsCalculator struct {
	metrics          Metrics
	tDigestEstimator *tdigest.TDigest
}

func NewMetricsCalculator() MetricsCalculator {
	metrics := Metrics{
		partitionDistribution: make(map[int]int),
	}
	return &metricsCalculator{
		tDigestEstimator: tdigest.NewWithCompression(1000),
		metrics:          metrics,
	}
}

func (metricsCalculator *metricsCalculator) AddRawMetric(rawMetricsData RawMetricsData) {
	metrics := metricsCalculator.metrics
	metrics.totalNoOfEvents++
	metrics.partitionDistribution[rawMetricsData.Partition]++
	latencyMetrics := metrics.latencyMetrics
	latency := rawMetricsData.Latency()
	latencyMetrics.total += latency
	metrics.latencyMetrics = latencyMetrics
	metricsCalculator.tDigestEstimator.Add(float64(latency), 1)
	metricsCalculator.metrics = metrics
}

func (metricsCalculator *metricsCalculator) GetMetrics() Metrics {
	metrics := metricsCalculator.metrics
	metrics.latencyMetrics.mean = time.Duration(float64(metrics.latencyMetrics.total) / float64(metrics.totalNoOfEvents))
	metrics.latencyMetrics.percentile95 = time.Duration(metricsCalculator.tDigestEstimator.Quantile(0.95))
	metrics.latencyMetrics.percentile99 = time.Duration(metricsCalculator.tDigestEstimator.Quantile(0.99))
	return metrics
}
