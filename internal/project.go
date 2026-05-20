package internal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	LabelManaged   = "opencodepod.managed"
	LabelProjectID = "opencodepod.project.id"
	LabelName      = "opencodepod.project.name"
	LabelGitRepo   = "opencodepod.project.git_repo"
	LabelImage     = "opencodepod.project.image"
)

type Project struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	GitRepo string `json:"git_repo,omitempty"`
	Status  string `json:"status"`
	SSHPort int    `json:"ssh_port"`
	WebPort int    `json:"web_port"`
	Image   string `json:"image"`
}

type CreateRequest struct {
	Name    string `json:"name"`
	GitRepo string `json:"git_repo,omitempty"`
	Image   string `json:"image,omitempty"`
}

func LabelsFromProject(p *Project) map[string]string {
	return map[string]string{
		LabelManaged:   "true",
		LabelProjectID: p.ID,
		LabelName:      p.Name,
		LabelGitRepo:   p.GitRepo,
		LabelImage:     p.Image,
	}
}

func ProjectFromLabels(id string, labels map[string]string) *Project {
	return &Project{
		ID:      id,
		Name:    labels[LabelName],
		GitRepo: labels[LabelGitRepo],
		Image:   labels[LabelImage],
	}
}

func ContainerName(id string) string {
	return fmt.Sprintf("cp-%s", id)
}

func ParsePort(ports []interface{}) int {
	// ports from nat.PortMap are strings
	for _, p := range ports {
		switch v := p.(type) {
		case string:
			parts := strings.Split(v, "/")
			if len(parts) > 0 {
				port, err := strconv.Atoi(parts[0])
				if err == nil {
					return port
				}
			}
		case map[string]interface{}:
			if hp, ok := v["HostPort"].(string); ok {
				port, err := strconv.Atoi(hp)
				if err == nil {
					return port
				}
			}
		}
	}
	return 0
}

func PrettyJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
