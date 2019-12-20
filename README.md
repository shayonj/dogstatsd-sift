## dogstatsd-sift

Small reverse proxy that sits between Datadog statsD agent and Datadog API. The compact proxy is used to filter/remove/strip unique tags before sending them to Datadog or remove metrics that match a certain condition. Common use case to strip/globalize `Host` on a `Metric`, instead of sending unique host/pod/container IDs to Datadog.

Currently, it only overrides the `Host` on every metric before sending it back to Datadog. More improvements to come soon where it will be possible to only override `Host` for a certain metrics (manageable via configuration).

Supports Datadog API V1.

## Setup

Run `dogstatsd-sift` on same network/location as the Datadog statsD agent and point DD URL to where its running. The exchange between Datadog agent and dogstatsd-sift happens over TCP (instead of UDP). This doesn't replace or touch any of the metrics ingestion that happens over UDP.

```yaml
# datadog.yml
...

dd_url: http://localhost:9000

...
```

## Logging

Logs for each request goes into `dogstatsd_sift_request.log` and `STDOUT`.


## Local

```
# go version go1.13.x


go mod download

go run main.go
```


