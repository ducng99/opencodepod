package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
)

type Mount struct {
	Source   string `json:"source" desc:"Absolute or relative host path to mount."`
	Target   string `json:"target" desc:"Path inside the container where the source is mounted."`
	ReadOnly bool   `json:"read_only" desc:"Whether the mount is read-only inside the container."`
}

type GitCredential struct {
	Username string `json:"username" desc:"Username for Git HTTP authentication."`
	Password string `json:"password" desc:"Password or PAT for Git HTTP authentication."`
}

type GitAuthConfig struct {
	SSHKey      string                     `json:"ssh_key" desc:"Inline SSH private key for Git authentication."`
	SSHKeyPath  string                     `json:"ssh_key_path" desc:"Container path where the SSH private key is written."`
	Credentials map[string]GitCredential `json:"credentials" desc:"Host-keyed username/password credentials for Git HTTP authentication."`
}

type GPGConfig struct {
	KeyID      string `json:"key_id" desc:"GPG key ID used for commit signing."`
	PrivateKey string `json:"private_key" desc:"Inline GPG private key for signing commits."`
}

type GitConfig struct {
	Auth      GitAuthConfig `json:"auth" desc:"SSH authentication settings for Git operations."`
	UserName  string        `json:"user_name" desc:"Git commit author name."`
	UserEmail string        `json:"user_email" desc:"Git commit author email."`
	GPG       GPGConfig     `json:"gpg" desc:"GPG signing configuration."`
}

type Config struct {
	ListenAddr   string            `json:"listen_addr" desc:"Address and port for the HTTP server to listen on (e.g., :8080)."`
	DefaultImage string            `json:"default_image" desc:"Default Docker image used for new project containers."`
	SSHPublicKey string            `json:"ssh_public_key" desc:"Public SSH key injected into containers for coder access."`
	Mounts       []Mount           `json:"mounts" desc:"Additional host paths to mount into containers."`
	Hosts        map[string]string `json:"hosts" desc:"Custom host entries added to container /etc/hosts."`
	Git          GitConfig         `json:"git" desc:"Git configuration including auth, user identity, and GPG signing."`
}

const defaultConfigPath = "config.json"

var filePlaceholderRe = regexp.MustCompile(`^\{file:(.+)\}$`)

func LoadConfig() (*Config, error) {
	return loadConfigFrom(defaultConfigPath)
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

	configDir := filepath.Dir(path)
	if err := expandPlaceholders(cfg, configDir); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func expandPlaceholders(v interface{}, configDir string) error {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return expandValue(rv, configDir)
}

func expandValue(v reflect.Value, configDir string) error {
	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if m := filePlaceholderRe.FindStringSubmatch(s); m != nil {
			filePath := m[1]
			if !filepath.IsAbs(filePath) {
				filePath = filepath.Join(configDir, filePath)
			}
			content, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			v.SetString(string(content))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if err := expandValue(v.Field(i), configDir); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := expandValue(v.Index(i), configDir); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if !v.IsNil() {
			if err := expandValue(v.Elem(), configDir); err != nil {
				return err
			}
		}
	}
	return nil
}
