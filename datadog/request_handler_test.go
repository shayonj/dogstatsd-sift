package datadog

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func examplePayload() *RequestSeriesPayload {
	return &RequestSeriesPayload{
		Series: []Metric{
			{
				Metric:         "test.metric",
				Points:         []DataPoint{{float64(1.0)}, {float64(2.0)}},
				Type:           "gauge",
				Host:           "i-myhost",
				Tags:           []string{"some:tag"},
				Unit:           "unit",
				SourceTypeName: "foo",
				Interval:       1,
			},
		},
	}
}

func TestDeflate(t *testing.T) {
	flatedResponse, err := json.Marshal(examplePayload())
	assert.Nil(t, err)

	var b bytes.Buffer
	w := zlib.NewWriter(&b)

	_, err = w.Write(flatedResponse)
	assert.Nil(t, err)

	err = w.Close()
	assert.Nil(t, err)

	result, err := deflate(b.Bytes())
	assert.Nil(t, err)

	assert.Equal(t, result, examplePayload())
}

func TestInflate(t *testing.T) {
	flatedResponse, err := json.Marshal(examplePayload())
	assert.Nil(t, err)

	var b bytes.Buffer
	w := zlib.NewWriter(&b)

	_, err = w.Write(flatedResponse)
	assert.Nil(t, err)

	err = w.Close()
	assert.Nil(t, err)

	bytes, err := inflate(examplePayload())
	assert.Nil(t, err)

	assert.Equal(t, bytes, b.Bytes())
}
