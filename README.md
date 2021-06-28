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

## Examples

## Collecting metrics

```shell
# Basic
$ kafka-perf monit --topic example --bootstrap-servers localhost:9092 > result.txt

# With TLS
$ kafka-perf monit --topic example --bootstrap-servers localhost:9092 -tls-mode TLS -tls-ca-cert /certs/ca.pem > result.txt

# With MTLS
$ kafka-perf monit --topic example --bootstrap-servers localhost:9092 -tls-mode MTLS -tls-cert /certs/cert.pem -tls-key /certs/key.pem -tls-ca-cert /certs/ca.pem > result.txt
```

## Generating report

```shell
# Text report
$ cat result.txt | kafka-perf report -type=text -timestamp-field=time > report.txt
$ cat report.txt
```

## Viewing collected metric information

```shell
# CSV Export
$ cat result.txt | kafka-perf encode -type=csv -timestamp-field=time > encoded-result.csv
$ head encoded-result.csv
```

## Advanced examples

## Profiling the utility

1. To profile just add `--profile` flag to any command you want to profile
2. To view the results run `go tool pprof -http=:1235 cpu.pprof`
3. Open url `http://localhost:1235` in your browser

## Usage

```shell
$ kafka-perf --help
Supported metrics
1. Latency - Mean, 95th percentile, 99th percentile
2. Total No of events
3. Partition distribution

Usage:
  kafka-perf [command]

Available Commands:
  encode      Encode results from one encoding to another
  help        Help about any command
  monit       Monitoring kafka events
  report      Outputs a report in the given format

Flags:
  -h, --help      help for kafka-perf
      --profile   profile this cli tool
  -v, --verbose   verbose output
      --version   version for kafka-perf

Use "kafka-perf [command] --help" for more information about a command.

$ kafka-perf monit --help
Monitoring kafka events

Usage:
  kafka-perf monit [flags]

Flags:
      --bootstrap-servers string   REQUIRED: The server(s) to connect to.
  -h, --help                       help for monit
      --tls-ca-cert string         CA cert file location. Eg. /certs/ca.pem. Required if tls-mode is TLS, MTLS
      --tls-cert string            certificate file location. Eg. /certs/cert.pem. Required if tls-mode is MTLS
      --tls-key string             key file location. Eg. /certs/key.pem. Required if tls-mode is MTLS
      --tls-mode string            Valid values are NONE,TLS,MTLS (default "NONE")
      --topic string               REQUIRED: The topic id to consume on

Global Flags:
      --profile   profile this cli tool
  -v, --verbose   verbose output

$ kafka-perf report --help
Outputs a report in the given format

Usage:
  kafka-perf report [flags]

Flags:
  -h, --help                     help for report
      --timestamp-field string   Field which has the unix timestamp. Eg 1617104831727
      --type string              Report type. Valid values are text (default "text")

Global Flags:
      --profile   profile this cli tool
  -v, --verbose   verbose output

$ kafka-perf encode --help
Encode results from one encoding to another

Usage:
  kafka-perf encode [flags]

Flags:
  -h, --help                     help for encode
      --timestamp-field string   Field which has the unix timestamp. Eg 1617104831727
      --type string              Report type. Valid values are csv (default "csv")

Global Flags:
      --profile   profile this cli tool
  -v, --verbose   verbose output
```