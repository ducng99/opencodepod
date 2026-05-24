package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

type Mount struct {
	Source   string `json:"source" desc:"Absolute or relative host path to mount."`
	Target   string `json:"target" desc:"Path inside the container where the source is mounted."`
	ReadOnly bool   `json:"read_only" desc:"Whether the mount is read-only inside the container."`
}

type GitCredential struct {
	Username string `json:"username" desc:"Username for Git HTTP authentication."`
	Password string `json:"password" desc:"Password or PAT for Git HTTP authentication." trim:"both"`
}

type GitAuthConfig struct {
	SSHKey      string                   `json:"ssh_key" desc:"Inline SSH private key for Git authentication."`
	SSHKeyPath  string                   `json:"ssh_key_path" desc:"Container path where the SSH private key is written."`
	Credentials map[string]GitCredential `json:"credentials" desc:"Host-keyed username/password credentials for Git HTTP authentication."`
}

type GPGConfig struct {
	KeyID      string `json:"key_id" desc:"GPG key ID used for commit signing."`
	PrivateKey string `json:"private_key" desc:"Inline GPG private key for signing commits."`
	Passphrase string `json:"passphrase" desc:"Passphrase for the GPG private key." trim:"both"`
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
		ListenAddr:   "127.0.0.1:8080",
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
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		return cfg, nil
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	configDir := filepath.Dir(path)
	if err := expandPlaceholders(cfg, configDir); err != nil {
		return nil, err
	}

	return cfg, nil
}

func expandPlaceholders(v any, configDir string) error {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return expandValue(rv, configDir, "")
}

func expandValue(v reflect.Value, configDir string, trim string) error {
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
			s = string(content)
		}
		if v.CanSet() {
			switch trim {
			case "prefix":
				s = strings.TrimLeft(s, " \t\n\r")
			case "suffix":
				s = strings.TrimRight(s, " \t\n\r")
			case "both":
				s = strings.TrimSpace(s)
			}
			v.SetString(s)
		}
	case reflect.Struct:
		for sf, field := range v.Fields() {
			if err := expandValue(field, configDir, sf.Tag.Get("trim")); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := expandValue(v.Index(i), configDir, ""); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			if !val.IsValid() {
				continue
			}
			copyVal := reflect.New(val.Type()).Elem()
			copyVal.Set(val)
			if err := expandValue(copyVal, configDir, ""); err != nil {
				return err
			}
			v.SetMapIndex(key, copyVal)
		}
	case reflect.Pointer:
		if !v.IsNil() {
			if err := expandValue(v.Elem(), configDir, ""); err != nil {
				return err
			}
		}
	}
	return nil
}
