package common

import (
	"gopkg.in/yaml.v2"
	"log"
)

func newConfig() *Config {
	return new(Config)
}

func (config *Config) loadFromYaml(yamlConfig []byte) *Config {
	err := yaml.Unmarshal(yamlConfig, config)
	if err != nil {
		log.Panicf("Invalid config file: %v", err)
	}

	return config
}
