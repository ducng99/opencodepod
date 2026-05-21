package internal

import (
	"testing"
)

func TestLabelsFromProject(t *testing.T) {
	p := &Project{
		ID:      "abc123",
		Name:    "myproject",
		GitRepo: "https://github.com/user/repo",
		Image:   "custom-opencode:latest",
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
	if p.GitRepo != "git@host:org/repo.git" {
		t.Errorf("expected git_repo, got %s", p.GitRepo)
	}
	if p.Image != "img:v2" {
		t.Errorf("expected image img:v2, got %s", p.Image)
	}
	if len(p.Volumes) != 2 {
		t.Errorf("expected 2 volumes, got %d", len(p.Volumes))
	}
	if p.Volumes[0] != VolumeName("xyz789") {
		t.Errorf("expected volume %s, got %s", VolumeName("xyz789"), p.Volumes[0])
	}
	if p.Volumes[1] != HomeVolumeName("xyz789") {
		t.Errorf("expected home volume %s, got %s", HomeVolumeName("xyz789"), p.Volumes[1])
	}
}

func TestVolumeName(t *testing.T) {
	if VolumeName("abc") != "cp-vol-abc" {
		t.Errorf("expected cp-vol-abc, got %s", VolumeName("abc"))
	}
}

func TestHomeVolumeName(t *testing.T) {
	if HomeVolumeName("abc") != "cp-vol-abc-home" {
		t.Errorf("expected cp-vol-abc-home, got %s", HomeVolumeName("abc"))
	}
}

func TestContainerName(t *testing.T) {
	if ContainerName("abc") != "cp-abc" {
		t.Errorf("expected cp-abc, got %s", ContainerName("abc"))
	}
}

func TestParsePort(t *testing.T) {
	cases := []struct {
		name     string
		ports    []interface{}
		expected int
	}{
		{"empty", []interface{}{}, 0},
		{"string", []interface{}{"8080/tcp"}, 8080},
		{"map", []interface{}{map[string]interface{}{"HostPort": "9090"}}, 9090},
		{"invalid string", []interface{}{"abc"}, 0},
		{"invalid map", []interface{}{map[string]interface{}{"HostPort": "bad"}}, 0},
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
