package common

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"fmt"
	"log"
	"strings"
)

// Config defines the structure of the yml file for the mu config
type Config struct {
	Environments []Environment `yaml:"environments,omitempty"`
	Service Service `yaml:"service,omitempty"`
}

// Service defines the service that will be created
type Service struct {
	DesiredCount int `yaml:"desiredCount,omitempty"`
	Pipeline ServicePipeline `yaml:"pipeline,omitempty"`
}

// ServicePipeline defines the service pipeline that will be created
type ServicePipeline struct {
}

// NewConfig create a new config object
func NewConfig() *Config {
	return &Config{}
}

// LoadFromFile loads config object from local file
func (config *Config) LoadFromFile(configFile string) {
	yamlConfig, err := ioutil.ReadFile( configFile )
	if err != nil {
		fmt.Printf("WARN: Unable to find config file: %v\n", err)
	} else {
		config.loadFromYaml(yamlConfig)
	}

}

func (config *Config) loadFromYaml(yamlConfig []byte)  *Config {
	err := yaml.Unmarshal(yamlConfig, config)
	if err != nil {
		log.Panicf("Invalid config file: %v", err)
	}

	return config
}

// GetEnvironment loads the environment by name from the config
func (config *Config) GetEnvironment(environmentName string) (*Environment, error) {

	for _, e := range config.Environments {
		if(strings.EqualFold(environmentName, e.Name)) {
			return &e, nil
		}
	}

	return nil, fmt.Errorf("Unable to find environment named '%s' in mu.yml",environmentName)
}
