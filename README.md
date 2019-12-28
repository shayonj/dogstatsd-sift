## dogstatsd-sift

Small reverse proxy that sits between Datadog statsD agent and Datadog API. The compact proxy is used to filter/remove/strip unique tags before sending them to Datadog or remove metrics that match a certain condition. Common use case to strip/globalize `Host` on a `Metric`, instead of sending unique host/pod/container IDs to Datadog.

Supports Datadog API V1.

## Setup

Run `dogstatsd-sift` on same network/location as the Datadog statsD agent and point DD URL to where its running. The exchange between Datadog agent and dogstatsd-sift happens over TCP (instead of UDP). This doesn't replace or touch any of the metrics ingestion that happens over UDP.

```yaml
# datadog.yml
...

dd_url: http://localhost:9000

...
```

## Run

```
go get github.com/shayonj/dogstatsd-sift

dogstatsd-sift --config-file="example.yml"
```

## Configuration

By passing a set of static configuration its easy to control the metrics/tags/host combination being sent out. Based on the static configuration ([example.yml](example.yml)) `dogstatsd-sift` modifies payload on the fly. Currently possible operations are:

- Selectively remove tags from specific metric(s)
- Selective override/disable `Host` from specific  metric(s)
- Globally override/disable `Host` from any metric(s)

## Logging

Logs for each request goes into `dogstatsd_sift_request.log` and `STDOUT`.

## Local

```
# go version go1.13.x


go mod download

go run main.go
```


