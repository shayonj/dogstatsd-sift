package datadog

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/shayonj/dogstatsd-sift/configuration"
	"github.com/sirupsen/logrus"
)

// OriginAPIEndpoint is the datadog endpoint
// to proxy request back to
const OriginAPIEndpoint = "/api/v1/series"

// ContentEncoding represents the value in which http
// request was received
const ContentEncoding = "deflate"

// ContentEncodingHeader represents the header value in which http
// request was received
const ContentEncodingHeader = "Content-Encoding"

// DataPoint is a tuple of [UNIX timestamp, value]. Values
// can be int too, hence float.
type DataPoint [2]float64

// Metric represents a collection of data points that we send/receive
// on /api/v1/series collection endpoint
type Metric struct {
	Metric         string      `json:"metric,omitempty"`
	Points         []DataPoint `json:"points,omitempty"`
	Type           string      `json:"type,omitempty"`
	Host           string      `json:"host,omitempty"`
	Tags           []string    `json:"tags,omitempty"`
	Unit           string      `json:"unit,omitempty"`
	SourceTypeName string      `json:"source_type_name,omitempty"`
	Interval       int         `json:"interval,omitempty"`
}

// RequestSeriesPayload collection from /api/v1/series
type RequestSeriesPayload struct {
	Series []Metric `json:"series,omitempty"`
}

// HandleRequest works on an http request to decode (from deflate), parse,
// modify and then encode back request in the way it was received, with the
// modified values, so it can be proxied back to the origin.
func HandleRequest(r *http.Request, log *logrus.Entry, cfg *configuration.Base) {
	if cfg == nil {
		log.Warn("No config found. Skipping request.")
	}

	if r.URL.Path != OriginAPIEndpoint {
		return
	}

	if r.Header.Get(ContentEncodingHeader) != ContentEncoding {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}

	enflatedPayload, err := deflate(body)
	if err != nil {
		log.Error(err)
		return
	}

	if err := mutate(enflatedPayload, cfg); err != nil {
		log.Error(err)
		return
	}

	innflatedPayload, err := inflate(enflatedPayload)
	if err != nil {
		log.Error(err)
		return
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(innflatedPayload))
	r.ContentLength = int64(len(innflatedPayload))
}

func deflate(body []byte) (reqPayload *RequestSeriesPayload, e error) {
	decompressedReader, err := zlib.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer decompressedReader.Close()

	enflated, err := ioutil.ReadAll(decompressedReader)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(enflated, &reqPayload); err != nil {
		return nil, err
	}

	return reqPayload, nil
}

func mutate(reqPayload *RequestSeriesPayload, cfg *configuration.Base) error {
	for i := 0; i < len(reqPayload.Series); i++ {
		metric := &reqPayload.Series[i]

		for j := 0; j < len(cfg.Metrics); j++ {
			mcfg := cfg.Metrics[j]

			if mcfg.RemoveMetric {
				reqPayload.Series = remove(reqPayload.Series, i)
				continue
			}

			if mcfg.Name == metric.Metric {
				if mcfg.RemoveHost {
					metric.Host = "dogstatsd-sift"
				}

				// TODO: filter tags based on mcfg.Tags
			}
		}

		if cfg.RemoveAllHost {
			metric.Host = "dogstatsd-sift"
		}
	}

	return nil
}

func remove(slice []Metric, s int) []Metric {
	return append(slice[:s], slice[s+1:]...)
}

func inflate(reqPayload *RequestSeriesPayload) ([]byte, error) {
	flatedResponse, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)

	_, err = w.Write(flatedResponse)
	if err != nil {
		return nil, err
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
