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

type Git struct {
	Repo   string `json:"repo,omitempty"`
	Branch string `json:"branch,omitempty"`
	Depth  int    `json:"depth,omitempty"`
}

type Project struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Git            Git      `json:"git"`
	Status         string   `json:"status"`
	SSHPort        int      `json:"ssh_port"`
	WebPort        int      `json:"web_port"`
	Volumes        []string `json:"volumes"`
	Image          string   `json:"image"`
	ContainerUser  string   `json:"container_user"`
}

type VolumeMount struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

type CreateRequest struct {
	Name  string `json:"name"`
	Git   Git    `json:"git"`
	Image string `json:"image,omitempty"`
}

type UpdateRequest struct {
	Name string `json:"name"`
}

func LabelsFromProject(p *Project) map[string]string {
	labels := map[string]string{
		LabelManaged:   "true",
		LabelProjectID: p.ID,
		LabelName:      p.Name,
		LabelGitRepo:   p.Git.Repo,
		LabelGitBranch: p.Git.Branch,
		LabelImage:     p.Image,
	}
	if p.Git.Depth > 0 {
		labels[LabelGitDepth] = strconv.Itoa(p.Git.Depth)
	}
	return labels
}

func ProjectFromLabels(id string, labels map[string]string) *Project {
	p := &Project{
		ID:      id,
		Name:    labels[LabelName],
		Git:     Git{Repo: labels[LabelGitRepo], Branch: labels[LabelGitBranch]},
		Image:   labels[LabelImage],
		Volumes: ProjectVolumes(id),
	}
	if d, err := strconv.Atoi(labels[LabelGitDepth]); err == nil {
		p.Git.Depth = d
	}
	return p
}

func ProjectVolumeMounts(id string, containerUser string) []VolumeMount {
	return []VolumeMount{
		{Name: WorkspacesVolumeName(id), Target: "/workspaces"},
		{Name: OpencodeSessionsVolumeName(id), Target: "/home/" + containerUser + "/.local/share/opencode"},
	}
}

func ProjectVolumes(id string) []string {
	return []string{
		WorkspacesVolumeName(id),
		OpencodeSessionsVolumeName(id),
	}
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
