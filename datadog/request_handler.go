package datadog

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io/ioutil"
	"net/http"

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
func HandleRequest(r *http.Request, log *logrus.Entry) {
	if r.URL.Path != OriginAPIEndpoint {
		return
	}
	if r.Header.Get(ContentEncodingHeader) != ContentEncoding {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)

		log.Error(err)
	}

	enflatedPayload, err := deflate(body)
	if err != nil {
		panic(err)
		log.Error(err)
		return
	}

	var reqPayload RequestSeriesPayload
	if err := mutate(&enflatedPayload, &reqPayload); err != nil {
		log.Error(err)
		panic(err)

		return
	}

	innflatedPayload, err := inflate(&reqPayload)
	if err != nil {
		panic(err)

		log.Error(err)
		return
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(innflatedPayload))
	r.ContentLength = int64(len(innflatedPayload))
}

func deflate(body []byte) ([]byte, error) {
	decompressedReader, err := zlib.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer decompressedReader.Close()

	enflated, err := ioutil.ReadAll(decompressedReader)
	if err != nil {
		return nil, err
	}

	return enflated, nil
}

func mutate(enflated *[]byte, reqPayload *RequestSeriesPayload) error {
	if err := json.Unmarshal(*enflated, &reqPayload); err != nil {
		return err
	}

	for i := 0; i < len(reqPayload.Series); i++ {
		metric := &reqPayload.Series[i]
		metric.Host = "dogstatsd-sift"
	}

	return nil
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
