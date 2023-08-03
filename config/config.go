package config

import (
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type Config struct {
	// Node ID
	ID     string `yaml:"id"`
	Server string `yaml:"server"`
	Local  string `yaml:"local"`
	Peer   Peer   `yaml:"peer"`
	// Maximum connection idle time (in seconds)
	Timeout int64 `yaml:"timeout"`
	// STUN server address
	Stun string `yaml:"stun"`
}

type Peer struct {
	ID   string `yaml:"id"`
	Bind string `yaml:"bind"`
}

func Default() Config {
	return Config{
		ID:      uuid.NewString(),
		Timeout: 120,
		Stun:    "stun.qq.com:3478",
	}
}

// Apply applies the given options to the config, returning the first error
// encountered (if any).
func (cfg *Config) Apply(opts ...Option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}

func (cfg *Config) Load(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, cfg)
}
