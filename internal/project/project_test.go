package project

import (
	"testing"
)

func TestLabelsFromProject(t *testing.T) {
	t.Parallel()
	p := &Project{
		ID:    "abc123",
		Name:  "myproject",
		Git:   Git{Repo: "https://github.com/user/repo"},
		Image: "custom-opencode:latest",
	}
	labels := LabelsFromProject(p)
	if labels[LabelManaged] != "true" {
		t.Errorf("expected managed=true, got %s", labels[LabelManaged])
	}
	if labels[LabelProjectID] != "abc123" {
		t.Errorf("expected project.id=abc123, got %s", labels[LabelProjectID])
	}
	if labels[LabelName] != "myproject" {
		t.Errorf("expected project.name=myproject, got %s", labels[LabelName])
	}
	if labels[LabelGitRepo] != "https://github.com/user/repo" {
		t.Errorf("expected project.git_repo, got %s", labels[LabelGitRepo])
	}
	if labels[LabelImage] != "custom-opencode:latest" {
		t.Errorf("expected project.image, got %s", labels[LabelImage])
	}
}

func TestProjectFromLabels(t *testing.T) {
	t.Parallel()
	labels := map[string]string{
		LabelProjectID: "xyz789",
		LabelName:      "test",
		LabelGitRepo:   "git@host:org/repo.git",
		LabelImage:     "img:v2",
	}
	p := ProjectFromLabels("xyz789", labels)
	if p.ID != "xyz789" {
		t.Errorf("expected id xyz789, got %s", p.ID)
	}
	if p.Name != "test" {
		t.Errorf("expected name test, got %s", p.Name)
	}
	if p.Git.Repo != "git@host:org/repo.git" {
		t.Errorf("expected git repo, got %s", p.Git.Repo)
	}
	if p.Image != "img:v2" {
		t.Errorf("expected image img:v2, got %s", p.Image)
	}
	if len(p.Volumes) != 2 {
		t.Errorf("expected 2 volumes, got %d", len(p.Volumes))
	}
	if p.Volumes[0] != WorkspacesVolumeName("xyz789") {
		t.Errorf("expected workspaces volume %s, got %s", WorkspacesVolumeName("xyz789"), p.Volumes[0])
	}
	if p.Volumes[1] != OpencodeSessionsVolumeName("xyz789") {
		t.Errorf("expected opencode sessions volume %s, got %s", OpencodeSessionsVolumeName("xyz789"), p.Volumes[1])
	}
}

func TestWorkspacesVolumeName(t *testing.T) {
	t.Parallel()
	if WorkspacesVolumeName("abc") != "cp-vol-abc-workspaces" {
		t.Errorf("expected cp-vol-abc-workspaces, got %s", WorkspacesVolumeName("abc"))
	}
}

func TestOpencodeSessionsVolumeName(t *testing.T) {
	t.Parallel()
	if OpencodeSessionsVolumeName("abc") != "cp-vol-abc-opencode" {
		t.Errorf("expected cp-vol-abc-opencode, got %s", OpencodeSessionsVolumeName("abc"))
	}
}

func TestContainerName(t *testing.T) {
	t.Parallel()
	if ContainerName("abc") != "cp-abc" {
		t.Errorf("expected cp-abc, got %s", ContainerName("abc"))
	}
}

func TestParsePort(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		ports    []any
		expected int
	}{
		{"empty", []any{}, 0},
		{"string", []any{"8080/tcp"}, 8080},
		{"map", []any{map[string]any{"HostPort": "9090"}}, 9090},
		{"invalid string", []any{"abc"}, 0},
		{"invalid map", []any{map[string]any{"HostPort": "bad"}}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParsePort(c.ports)
			if got != c.expected {
				t.Errorf("expected %d, got %d", c.expected, got)
			}
		})
	}
}
