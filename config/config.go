package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type HAConfig struct {
	Token string `yaml:"token"`
	URL   string `yaml:"url"`
}

type AppConfig struct {
	HAConfig HAConfig `yaml:"ha"`
}

func LoadConfig(file string) (*AppConfig, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	config := &AppConfig{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
