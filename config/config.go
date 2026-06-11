package config

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig   `toml:"server"`
	Database DatabaseConfig `toml:"database"`
	Security SecurityConfig `toml:"security"`
	Engine   EngineConfig   `toml:"engine"`
	UI       UIConfig       `toml:"ui"`
}

type ServerConfig struct {
	Listen  string `toml:"listen"`
	Timeout string `toml:"timeout"`
}

type DatabaseConfig struct {
	Path string `toml:"path"`
}

type SecurityConfig struct {
	MaxTimeout    int `toml:"max_timeout"`
	MaxMemory     int `toml:"max_memory"`
	MaxMatrixSize int `toml:"max_matrix_size"`
	MaxTensorNdim int `toml:"max_tensor_ndim"`
}

type EngineConfig struct {
	Edition string `toml:"edition"`
}

type UIConfig struct {
	Theme     string `toml:"theme"`
	AutoSave  bool   `toml:"auto_save"`
	FontSize  int    `toml:"font_size"`
	TabSize   int    `toml:"tab_size"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg.applyDefaults()
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}
	cfg.applyDefaults()
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Listen == "" {
		c.Server.Listen = ":8080"
	}
	if c.Server.Timeout == "" {
		c.Server.Timeout = "30s"
	}
	if c.Database.Path == "" {
		c.Database.Path = "./data/koi.db"
	}
	if c.Security.MaxTimeout == 0 {
		c.Security.MaxTimeout = 60
	}
	if c.Security.MaxMemory == 0 {
		c.Security.MaxMemory = 1073741824
	}
	if c.Security.MaxMatrixSize == 0 {
		c.Security.MaxMatrixSize = 10000
	}
	if c.Security.MaxTensorNdim == 0 {
		c.Security.MaxTensorNdim = 8
	}
	if c.Engine.Edition == "" {
		c.Engine.Edition = "full"
	}
	if c.UI.Theme == "" {
		c.UI.Theme = "dark"
	}
	if c.UI.FontSize == 0 {
		c.UI.FontSize = 14
	}
	if c.UI.TabSize == 0 {
		c.UI.TabSize = 4
	}
}

func (c *Config) GetTimeout() time.Duration {
	d, err := time.ParseDuration(c.Server.Timeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (c *Config) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}
