package config

type Option func(cfg *Config) error

func Timeout(s int64) Option {
	return func(cfg *Config) error {
		cfg.Timeout = s
		return nil
	}
}

func Stun(server string) Option {
	return func(cfg *Config) error {
		cfg.Stun = server
		return nil
	}
}
