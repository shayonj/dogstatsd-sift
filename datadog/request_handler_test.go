package datadog

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"testing"

	"github.com/shayonj/dogstatsd-sift/configuration"
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
				Tags:           []string{"some_tag: true", "foo_bar:false"},
				Unit:           "unit",
				SourceTypeName: "foo",
				Interval:       1,
			},
			{
				Metric:         "request.200",
				Points:         []DataPoint{{float64(1.0)}, {float64(2.0)}},
				Type:           "gauge",
				Host:           "i-myhost",
				Tags:           []string{"status:200", "foo_bar:false"},
				Unit:           "unit",
				SourceTypeName: "foo",
				Interval:       1,
			},
		},
	}
}

func exampleConfig() *configuration.Base {
	return &configuration.Base{
		Port: 9000,
		Metrics: []configuration.Metrics{
			{
				Name:         "test.metric",
				RemoveMetric: false,
				RemoveTags:   []string{},
				RemoveHost:   true,
			},
			{
				Name:         "request.200",
				RemoveMetric: false,
				RemoveTags:   []string{},
				RemoveHost:   false,
			},
		},
		RemoveAllHost: false,
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

	assert.Equal(t, examplePayload(), result)
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

	assert.Equal(t, b.Bytes(), bytes)
}

func TestMutateRemoveAllHost(t *testing.T) {
	ecfg := exampleConfig()

	p := examplePayload()
	err := mutate(p, ecfg)
	assert.Nil(t, err)

	metric := p.Series[0]
	assert.Equal(t, "test.metric", metric.Metric)
	assert.Equal(t, "dogstatsd-sift", metric.Host)
	assert.Equal(t, []DataPoint{DataPoint{1, 0}, DataPoint{2, 0}}, metric.Points)
}

func TestMutateRemoveHostForSingleMetric(t *testing.T) {
	p := examplePayload()

	err := mutate(p, exampleConfig())
	assert.Nil(t, err)

	testMetric := p.Series[0]
	requestMetric := p.Series[1]

	tests := []struct {
		in  string
		out string
	}{
		{in: "test.metric", out: testMetric.Metric},
		{in: "dogstatsd-sift", out: testMetric.Host},
		{in: "request.200", out: requestMetric.Metric},
		{in: "i-myhost", out: requestMetric.Host},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.in, tt.out)
		})
	}

	assert.Equal(t, []DataPoint{DataPoint{1, 0}, DataPoint{2, 0}}, testMetric.Points)
	assert.Equal(t, []DataPoint{DataPoint{1, 0}, DataPoint{2, 0}}, requestMetric.Points)
}

func TestMutateRemoveSingleMetric(t *testing.T) {
	ecfg := exampleConfig()
	ecfg.Metrics[1].RemoveMetric = true

	p := examplePayload()

	err := mutate(p, ecfg)
	assert.Nil(t, err)

	assert.Len(t, p.Series, 1)
	testMetric := p.Series[0]

	tests := []struct {
		in  string
		out string
	}{
		{in: "test.metric", out: testMetric.Metric},
		{in: "dogstatsd-sift", out: testMetric.Host},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.in, tt.out)
		})
	}

	assert.Equal(t, []DataPoint{DataPoint{1, 0}, DataPoint{2, 0}}, testMetric.Points)
}

func TestMutateRemoveTagsFromSingleMetric(t *testing.T) {
	ecfg := exampleConfig()
	ecfg.Metrics[1].RemoveTags = []string{"status:200"}

	p := examplePayload()

	err := mutate(p, ecfg)
	assert.Nil(t, err)

	assert.Len(t, p.Series, 2)

	testMetric := p.Series[0]
	requestMetric := p.Series[1]

	tests := []struct {
		in  string
		out string
	}{
		{in: "test.metric", out: testMetric.Metric},
		{in: "dogstatsd-sift", out: testMetric.Host},
		{in: "request.200", out: requestMetric.Metric},
		{in: "i-myhost", out: requestMetric.Host},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.in, tt.out)
		})
	}

	testTags := []struct {
		in  []string
		out []string
	}{
		{in: []string{"some_tag: true", "foo_bar:false"}, out: testMetric.Tags},
		{in: []string{"foo_bar:false"}, out: requestMetric.Tags},
	}

	for _, tt := range testTags {
		t.Run("testing tags", func(t *testing.T) {
			assert.Equal(t, tt.in, tt.out)
		})
	}
}
