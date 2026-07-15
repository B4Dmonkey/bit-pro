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
	f, err := os.Create(configFileName)
	if err != nil {
		return fmt.Errorf("creating %s: %w", configFileName, err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("writing %s: %w", configFileName, err)
	}
	return nil
}
