package config

import (
	"GAServer/config"
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Base         config.ServiceConfig `json:"config"`
	DB           map[string]string    `json:"db"`
	DesignConfig map[string]string    `json:"design"`
	Ver          string               `json:"ver"`
}

var appConfig *Config

func LoadConfig(confPath string) (*Config, error) {
	if data, err := ioutil.ReadFile(confPath); err != nil {
		return nil, err
	} else {
		var conf = &Config{}
		err := json.Unmarshal(data, conf)
		appConfig = conf
		return conf, err
	}

}

func GetAppConf() *Config {
	return appConfig
}
