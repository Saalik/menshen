package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port       string        `yaml:"port"`
	TTL        time.Duration `yaml:"ttl"`
	RateLimits RateLimits    `yaml:"rate_limits"`
	LogLevel   string        `yaml:"log_level"`
}

type RateLimits struct {
	Global int `yaml:"global"`
	Repo   int `yaml:"repo"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
