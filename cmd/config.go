package cmd

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const configFileName = ".bit/config.toml"

type Config struct {
	Prefix string `toml:"prefix"`
}

func loadConfig() (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(configFileName, &cfg); err != nil {
		return nil, fmt.Errorf("reading %s: %w", configFileName, err)
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", configFileName, err)
	}
	if err := os.WriteFile(configFileName, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", configFileName, err)
	}
	return nil
}
