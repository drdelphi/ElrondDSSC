package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/DrDelphi/ElrondDSSC/data"
)

// NewConfig - reads the application configuration from the provided path
// and returns an AppConfig struct or an error if something goes wrong
func NewConfig(configPath string) (*data.AppConfig, error) {
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := &data.AppConfig{}
	err = json.Unmarshal(bytes, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
