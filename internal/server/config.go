package server

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var configYAML []byte

type Config struct {
	Auth struct {
		RefreshTokenTtlSeconds uint32 `yaml:"refreshTokenTtlSeconds"`
		TokenTtlSeconds        uint32 `yaml:"tokenTtlSeconds"`
	} `yaml:"auth"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(configYAML, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
