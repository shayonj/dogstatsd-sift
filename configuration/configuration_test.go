package configuration

import (
	"reflect"
	"testing"
)

func TestParseInValidFile(t *testing.T) {
	_, err := Parse("test_data/bad_config.yml")

	if err == nil {
		t.Errorf("No error raised. Expected: %s", err)
	}
}

func TestParseValidFile(t *testing.T) {
	cfg, err := Parse("test_data/good_config.yml")

	if err != nil {
		t.Errorf("Exception raised %s", err)
	}

	expectedCfg := &Base{
		Port: 9000,
		Metrics: []Metrics{
			{
				Name:         "request.200",
				RemoveMetric: true,
				RemoveTags:   []string{"some-tags"},
				RemoveHost:   true,
			},
		},
		RemoveAllHost: true,
	}

	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Errorf("Structs didn't match. Recieved: %v. Expected: %vs", cfg, expectedCfg)
	}
}

func TestParseNoValidFile(t *testing.T) {
	_, err := Parse("foo-bar.yml")

	if err == nil {
		t.Error("No error raised")
	}
	if err.Error() != "open foo-bar.yml: no such file or directory" {
		t.Errorf("Unexpected exception raised. Expecting a no file found error %s", err)
	}
}
