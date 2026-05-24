package project

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
	LabelGitBranch = "opencodepod.project.git_branch"
	LabelGitDepth  = "opencodepod.project.git_depth"
	LabelImage     = "opencodepod.project.image"
)

type Project struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	GitRepo   string   `json:"git_repo,omitempty"`
	GitBranch string   `json:"git_branch,omitempty"`
	GitDepth  int      `json:"git_depth,omitempty"`
	Status    string   `json:"status"`
	SSHPort   int      `json:"ssh_port"`
	WebPort   int      `json:"web_port"`
	Volumes   []string `json:"volumes"`
	Image     string   `json:"image"`
}

type VolumeMount struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

type CreateRequest struct {
	Name      string `json:"name"`
	GitRepo   string `json:"git_repo,omitempty"`
	GitBranch string `json:"git_branch,omitempty"`
	GitDepth  int    `json:"git_depth,omitempty"`
	Image     string `json:"image,omitempty"`
}

type UpdateRequest struct {
	Name string `json:"name"`
}

func LabelsFromProject(p *Project) map[string]string {
	labels := map[string]string{
		LabelManaged:   "true",
		LabelProjectID: p.ID,
		LabelName:      p.Name,
		LabelGitRepo:   p.GitRepo,
		LabelGitBranch: p.GitBranch,
		LabelImage:     p.Image,
	}
	if p.GitDepth > 0 {
		labels[LabelGitDepth] = strconv.Itoa(p.GitDepth)
	}
	return labels
}

func ProjectFromLabels(id string, labels map[string]string) *Project {
	p := &Project{
		ID:        id,
		Name:      labels[LabelName],
		GitRepo:   labels[LabelGitRepo],
		GitBranch: labels[LabelGitBranch],
		Image:     labels[LabelImage],
		Volumes:   ProjectVolumes(id),
	}
	if d, err := strconv.Atoi(labels[LabelGitDepth]); err == nil {
		p.GitDepth = d
	}
	return p
}

func ProjectVolumeMounts(id string) []VolumeMount {
	return []VolumeMount{
		{Name: WorkspacesVolumeName(id), Target: "/workspaces"},
		{Name: OpencodeSessionsVolumeName(id), Target: "/home/coder/.local/share/opencode"},
	}
}

func ProjectVolumes(id string) []string {
	mounts := ProjectVolumeMounts(id)
	names := make([]string, len(mounts))
	for i, m := range mounts {
		names[i] = m.Name
	}
	return names
}

func WorkspacesVolumeName(id string) string {
	return fmt.Sprintf("cp-vol-%s-workspaces", id)
}

func OpencodeSessionsVolumeName(id string) string {
	return fmt.Sprintf("cp-vol-%s-opencode", id)
}

func ContainerName(id string) string {
	return fmt.Sprintf("cp-%s", id)
}

func ParsePort(ports []any) int {
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
		case map[string]any:
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

func PrettyJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
