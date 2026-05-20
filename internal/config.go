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

type GitAuthConfig struct {
	SSHKey     string `json:"ssh_key"`
	SSHKeyPath string `json:"ssh_key_path"`
}

type GitConfig struct {
	Auth GitAuthConfig `json:"auth"`
}

type Config struct {
	ListenAddr   string    `json:"listen_addr"`
	DefaultImage string    `json:"default_image"`
	SSHPublicKey string    `json:"ssh_public_key"`
	Mounts       []Mount   `json:"mounts"`
	Git          GitConfig `json:"git"`
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
		Git: GitConfig{
			Auth: GitAuthConfig{
				SSHKeyPath: "/home/coder/.ssh/id_ed25519",
			},
		},
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
