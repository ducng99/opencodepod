package internal

import (
	"encoding/json"
	"os"
)

type Mount struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"read_only"`
}

type Config struct {
	ListenAddr   string  `json:"listen_addr"`
	DefaultImage string  `json:"default_image"`
	SSHPublicKey string  `json:"ssh_public_key"`
	Mounts       []Mount `json:"mounts"`
}

const defaultConfigPath = "config.json"

func LoadConfig() *Config {
	cfg, _ := loadConfigFrom(defaultConfigPath)
	return cfg
}

func loadConfigFrom(path string) (*Config, error) {
	cfg := &Config{
		ListenAddr:   ":8080",
		DefaultImage: "ghcr.io/ducng99/opencodepod-client:latest",
		SSHPublicKey: "",
		Mounts:       []Mount{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
