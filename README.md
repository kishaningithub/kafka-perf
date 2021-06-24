# kafka-perf
Get performance metrics based on kafka events

![build workflow](https://github.com/kishaningithub/kafka-perf/actions/workflows/build.yml/badge.svg)

Tail kafka avro topic data without confluent schema registry overhead

This expects the data to be written in [Object Container File (OCF)](https://avro.apache.org/docs/current/spec.html#Object+Container+Files) format

## Installation

```shell
$ brew tap kishaningithub/tap
$ brew install kafka-perf
```

## Upgrading

```shell
$ brew upgrade kafka-perf
```

## Usage

```shell
$ kafka-perf -help
Usage of kafka-perf:
  -bootstrap-servers string
    	REQUIRED: The server(s) to connect to.
  -tls-ca-cert string
    	CA cert file location. Eg. /certs/ca.pem. Required if tls-mode is TLS, MTLS
  -tls-cert string
    	certificate file location. Eg. /certs/cert.pem. Required if tls-mode is MTLS
  -tls-key string
    	key file location. Eg. /certs/key.pem. Required if tls-mode is MTLS
  -tls-mode string
    	Valid values are NONE,TLS,MTLS (default "NONE")
  -topic string
    	REQUIRED: The topic id to consume on.
```

## Examples

### Basic usage

```shell
$ kafka-perf --topic example --bootstrap-servers localhost:9092

Schema
=====
{"fields":[{"name":"time","type":"long"},{"default":"","description":"Process id","name":"process_id","type":"string"}],"name":"example","namespace":"com.example","type":"record","version":1}
Data
=====
{"time":1617104831727, "process_id":"ID1"}
{"time":1717104831727, "process_id":"ID2"}

Schema
=====
{"fields":[{"name":"time","type":"long"}],"name":"example","namespace":"com.example","type":"record","version":2}
Data
=====
{"time":1817104831727}
{"time":1917104831727}
```

### With TLS

```shell
$ kafka-perf --topic example --bootstrap-servers localhost:9092 -tls-mode TLS -tls-ca-cert /certs/ca.pem

Schema
=====
{"fields":[{"name":"time","type":"long"},{"default":"","description":"Process id","name":"process_id","type":"string"}],"name":"example","namespace":"com.example","type":"record","version":1}
Data
=====
{"time":1617104831727, "process_id":"ID1"}
{"time":1717104831727, "process_id":"ID2"}

Schema
=====
{"fields":[{"name":"time","type":"long"}],"name":"example","namespace":"com.example","type":"record","version":2}
Data
=====
{"time":1817104831727}
{"time":1917104831727}
```

### With MTLS

```shell
$ kafka-perf --topic example --bootstrap-servers localhost:9092 -tls-mode MTLS -tls-cert /certs/cert.pem -tls-key /certs/key.pem -tls-ca-cert /certs/ca.pem

Schema
=====
{"fields":[{"name":"time","type":"long"},{"default":"","description":"Process id","name":"process_id","type":"string"}],"name":"example","namespace":"com.example","type":"record","version":1}
Data
=====
{"time":1617104831727, "process_id":"ID1"}
{"time":1717104831727, "process_id":"ID2"}

Schema
=====
{"fields":[{"name":"time","type":"long"}],"name":"example","namespace":"com.example","type":"record","version":2}
Data
=====
{"time":1817104831727}
{"time":1917104831727}
```
