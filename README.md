# kafka-perf
Get performance metrics based on kafka events

![build workflow](https://github.com/kishaningithub/kafka-perf/actions/workflows/build.yml/badge.svg)

This currently expects the data to be written in [Object Container File (OCF)](https://avro.apache.org/docs/current/spec.html#Object+Container+Files) format

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
$ kafka-perf monit -help
Usage of monit:
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
    	
$ kafka-perf report -help
Usage of report:
  -timestamp-field string
    	Field which has the unix timestamp. Eg 1617104831727
  -type string
    	Report type. Valid values are text (default "text")
```

## Examples

### Basic usage

```shell
# Collecting metrics
$ kafka-perf monit --topic example --bootstrap-servers localhost:9092 > result.txt

# Querying metrics
$ cat result.txt | kafka-perf report -type=text -timestamp-field=time

```

### With TLS

```shell
# Collecting metrics
$ kafka-perf monit --topic example --bootstrap-servers localhost:9092 -tls-mode TLS -tls-ca-cert /certs/ca.pem  > result.txt

# Querying metrics
$ cat result.txt | kafka-perf report -type=text -timestamp-field=time
```

### With MTLS

```shell
# Collecting metrics
$ kafka-perf monit --topic example --bootstrap-servers localhost:9092 -tls-mode MTLS -tls-cert /certs/cert.pem -tls-key /certs/key.pem -tls-ca-cert /certs/ca.pem

# Querying metrics
$ cat result.txt | kafka-perf report -type=text -timestamp-field=time
```
