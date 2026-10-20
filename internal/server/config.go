package server

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var configYAML []byte

type Config struct {
	Build struct {
		BuildTime       string
		GenerateSecrets bool
	}
	Auth struct {
		RefreshTokenTtlSeconds uint32 `yaml:"refreshTokenTtlSeconds"`
		TokenTtlSeconds        uint32 `yaml:"tokenTtlSeconds"`
		SecretKeyPath          string `yaml:"secretKeyPath"`
		PublicKeyPath          string `yaml:"publicKeyPath"`
		Crypto                 struct {
			Algorithm string `yaml:"algorithm"`
			KeySize   int    `yaml:"keySize"`
		} `yaml:"crypto"`
	} `yaml:"auth"`
}

func LoadConfig(buildTime string, generateSecrets bool) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(configYAML, &cfg)
	if err != nil {
		return nil, err
	}
	cfg.Build.BuildTime = buildTime
	cfg.Build.GenerateSecrets = generateSecrets
	return &cfg, nil
}
