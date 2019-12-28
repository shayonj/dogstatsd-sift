package configuration

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Metrics represents the static configuration of a metric entity from Base
type Metrics struct {
	Name         string   `yaml:"name"`
	RemoveMetric bool     `yaml:"remove_metric"`
	RemoveTags   []string `yaml:"remove_tags"`
	RemoveHost   bool     `yaml:"remove_host"`
}

// Base represents the static configuration read from yaml file
type Base struct {
	Metrics       []Metrics `yaml:"metrics"`
	RemoveAllHost bool      `yaml:"remove_all_host"`
	Port          int       `yaml:"port"`
}

// Parse takes the config location, parses the yaml and
// returns Config struct represenation of the same
func Parse(fileLocation string) (cfg *Base, e error) {
	f, err := os.Open(fileLocation)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
