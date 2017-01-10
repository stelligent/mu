package common

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"fmt"
	"log"
)

// NewConfig create a new config object
func NewConfig() *Config {
	return new(Config)
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


