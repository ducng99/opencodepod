package internal

import (
	"os"
)

type Config struct {
	ListenAddr   string
	DefaultImage string
	SSHPublicKey string
}

func LoadConfig() *Config {
	cfg := &Config{
		ListenAddr:   getEnv("APP_LISTEN", ":8080"),
		DefaultImage: getEnv("DEFAULT_IMAGE", "ghcr.io/ducng99/opencodepod-client:latest"),
		SSHPublicKey: getEnv("APP_SSH_PUBLIC_KEY", ""),
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
