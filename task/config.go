package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Prefix string `toml:"prefix"`
}

func (s *Store) ConfigPath() string {
	return filepath.Join(s.root, configFileName)
}

func (s *Store) Config() (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(s.ConfigPath(), &cfg); err != nil {
		return nil, fmt.Errorf("reading %s: %w", s.ConfigPath(), err)
	}
	return &cfg, nil
}

func (s *Store) SaveConfig(cfg *Config) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", s.ConfigPath(), err)
	}
	if err := os.MkdirAll(s.root, dirMode); err != nil {
		return fmt.Errorf("creating %s: %w", s.root, err)
	}
	if err := os.WriteFile(s.ConfigPath(), data, fileMode); err != nil {
		return fmt.Errorf("writing %s: %w", s.ConfigPath(), err)
	}
	return nil
}
