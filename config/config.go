package config

import (
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type Config struct {
	// Node ID.
	ID string `yaml:"id"`
	// Peer exchange server.
	// Redis only for now, eg: redis://....
	Server string `yaml:"server"`
	// Local is the destination address where traffic will be sent.
	Local string `yaml:"local"`
	Peer  Peer   `yaml:"peer,omitempty"`
	// Maximum connection idle time (in seconds).
	Timeout int64 `yaml:"timeout,omitempty"`
	// STUN server address.
	Stun string `yaml:"stun,omitempty"`
}

type Peer struct {
	ID string `yaml:"id"`
	// Bind is the listen address for receiving traffic.
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

// Load config from the given file path.
func (cfg *Config) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

// Save config to the give file path.
func (cfg *Config) Save(path string, flag int) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, flag, 0666)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
