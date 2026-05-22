package config

import "time"

type AppConfig struct {
	Server ServerConfig `yaml:"server"`
	Store  StoreConfig  `yaml:"store"`
}

type ServerConfig struct {
	Port     int           `yaml:"port"`
	Timeouts TimeoutConfig `yaml:"timeouts"`
}

type TimeoutConfig struct {
	Read     time.Duration `yaml:"read"`
	Write    time.Duration `yaml:"write"`
	Idle     time.Duration `yaml:"idle"`
	Shutdown time.Duration `yaml:"shutdown"`
}

type StoreConfig struct {
	Ttl time.Duration `yaml:"time_to_leave"`
}
