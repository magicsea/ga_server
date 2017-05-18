package main

import (
	"GAServer/config"
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Base config.ServiceConfig `json:"config"`
	Ver  string               `json:"ver"`
}

func LoadConfig(confPath string) (*Config, error) {
	if data, err := ioutil.ReadFile(confPath); err != nil {
		return nil, err
	} else {
		var conf = &Config{}
		err := json.Unmarshal(data, conf)
		return conf, err
	}

}
