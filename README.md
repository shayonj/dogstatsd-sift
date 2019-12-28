## dogstatsd-sift

Small reverse proxy that sits between Datadog statsD agent and Datadog API. The compact proxy is used to filter/remove/strip unique tags or unnecessary metrics on the fly, before sending them to Datadog. Common use case to strip/globalize `Host` on a `Metric`, instead of sending unique host/pod/container IDs to Datadog.

Dogstatsd-sift works best when used datadog-agent as a proxy. More on setup below.

## Requirements

### Runtime
Datadog agent: 6.x or 7.x

### Local
Go: 1.9+

## Overhead

- CPU: <1%
- Mem/RSS: ~40MB

Note: Works Datadog API V1 (`/api/v1/series`) payload type.

## Setup

Run `dogstatsd-sift` on same network/location as the Datadog statsD agent and point DD URL to where its running. The exchange between Datadog agent and dogstatsd-sift happens over TCP (instead of UDP). This doesn't replace or touch any of the metrics ingestion that happens over UDP.

In this setup, Datadog agent continues to receive data over UDP per usual, but instead of sending the data directly to Datadog API, it is proxied via `dogstatsd-sift`. Hence, the agent enables a proxy setup.

```yaml
# datadog.yml
...

dd_url: http://localhost:9000 (or dogstatsd-sift is running on)

...
```

## Run
```
dogstatsd-sift --config-file="example.yml"
```

Binary from releases: https://github.com/shayonj/dogstatsd-sift/releases

or if `$GOPATH/bin` is added to `PATH`

```
go get github.com/shayonj/dogstatsd-sift
```

## Configuration

By passing a set of static configuration its easy to control the metrics/tags/host combination being sent out. Based on the static configuration ([example.yml](example.yml)) `dogstatsd-sift` modifies payload on the fly. Currently possible operations/tasks include:

- Selectively remove tags from specified metric(s)
- Selectively override/disable `Host` from specified metric(s)
- Globally override/disable `Host` from any metric(s)

## Logging

Logs for each request goes into `dogstatsd_sift_request.log` and `STDOUT`.

## Local

```
# go version go1.13.x


go mod download

go run main.go
```

## Release

```
export GITHUB_TOKEN=`YOUR_GH_TOKEN`

git tag -a v0.1.0 -m "First release"
git push origin v0.1.0

goreleaser
```
