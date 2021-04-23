package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

func LoadConfigJson(filename string, config interface{}) error {
	if filename == "" {
		return errors.New("invalid config file path")
	}
	if config == nil {
		return errors.New("nil reference on config struct")
	}

	configData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(configData, config)
}
