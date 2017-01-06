package common

// Config defines the structure of the yml file for the mu config
type Config struct {
}

// LoadConfig creates a new config object and loads from local file
func LoadConfig() *Config {
	return &Config {
	}
}
