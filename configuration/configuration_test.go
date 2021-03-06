package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInValidFile(t *testing.T) {
	_, err := Parse("test_data/bad_config.yml")
	assert.NotNil(t, err)
}

func TestParseValidFile(t *testing.T) {
	cfg, err := Parse("test_data/good_config.yml")
	assert.Nil(t, err)

	expectedCfg := &Base{
		Port: 9000,
		Metrics: []Metrics{
			{
				Name:         "request.200",
				RemoveMetric: true,
				RemoveTags:   []string{"some_tags:true"},
				RemoveHost:   true,
			},
		},
		RemoveAllHost: true,
	}

	assert.Equal(t, cfg, expectedCfg)
}

func TestParseNoValidFile(t *testing.T) {
	_, err := Parse("foo-bar.yml")

	assert.NotNil(t, err)

	if err.Error() != "open foo-bar.yml: no such file or directory" {
		t.Errorf("Unexpected exception raised. Expecting a no file found error %s", err)
	}
}
