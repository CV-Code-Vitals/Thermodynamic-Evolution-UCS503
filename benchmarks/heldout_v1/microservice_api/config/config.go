package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// Neg_Control_LoadConfigSafe handles errors appropriately
func LoadConfigSafe(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// DEF-HOLDOUT-01: json.Unmarshal error silently discarded
func LoadConfigUnchecked(data []byte) *Config {
	var cfg Config
	_ = json.Unmarshal(data, &cfg) // UNHANDLED_ERROR: discarded error on unmarshal
	return &cfg
}
