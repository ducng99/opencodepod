package internal

import (
	"os"
)

type Config struct {
	ListenAddr           string
	DefaultImage         string
	SSHPublicKey         string
	OpenCodeConfigPath   string
	OpenCodeConfigTarget string
}

func LoadConfig() *Config {
	cfg := &Config{
		ListenAddr:           getEnv("APP_LISTEN", ":8080"),
		DefaultImage:         getEnv("DEFAULT_IMAGE", "ghcr.io/ducng99/opencodepod-client:latest"),
		SSHPublicKey:         getEnv("APP_SSH_PUBLIC_KEY", ""),
		OpenCodeConfigPath:   getEnv("OPENCODE_CONFIG_PATH", ""),
		OpenCodeConfigTarget: getEnv("OPENCODE_CONFIG_TARGET", ""),
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
