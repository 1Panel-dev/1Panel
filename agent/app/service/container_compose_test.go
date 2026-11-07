package service

import (
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/docker/docker/api/types/container"
)

func TestLoadComposeProjectNameForLog(t *testing.T) {
	list := []container.Summary{
		{Labels: map[string]string{
			composeProjectLabel: "openclaw",
			composeConfigLabel:  "/opt/1panel/apps/openclaw/OpenClaw/docker-compose.yml",
		}},
		{Labels: map[string]string{
			composeProjectLabel: "openclaw",
			composeConfigLabel:  "/opt/1panel/apps/openclaw/OpenClaw/docker-compose.yml",
		}},
		{Labels: map[string]string{
			composeProjectLabel: "demo",
			composeConfigLabel:  "/opt/1panel/docker/compose/demo/docker-compose.yml,/opt/1panel/docker/compose/demo/override.yml",
		}},
	}
	tests := []struct {
		name        string
		inputName   string
		composePath string
		want        string
	}{
		{name: "display name differs in case", inputName: "OpenClaw", composePath: "/opt/1panel/apps/openclaw/OpenClaw/docker-compose.yml", want: "openclaw"},
		{name: "project name matches", inputName: "demo", composePath: "/opt/1panel/docker/compose/demo/docker-compose.yml", want: "demo"},
		{name: "multiple compose files", inputName: "Demo", composePath: "/opt/1panel/docker/compose/demo/docker-compose.yml", want: "demo"},
		{name: "no matching container", inputName: "other", composePath: "/opt/1panel/apps/other/docker-compose.yml", want: "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadComposeProjectNameForLog(list, tt.inputName, tt.composePath)
			if err != nil {
				t.Fatalf("loadComposeProjectNameForLog(%q, %q) returned error: %v", tt.inputName, tt.composePath, err)
			}
			if got != tt.want {
				t.Fatalf("loadComposeProjectNameForLog(%q, %q) = %q, want %q", tt.inputName, tt.composePath, got, tt.want)
			}
		})
	}
}

func TestLoadComposeProjectNameForLogSharedComposeFile(t *testing.T) {
	re.Init()
	const composePath = "/opt/1panel/apps/openclaw/OpenClaw/docker-compose.yml"
	openclaw := container.Summary{Labels: map[string]string{
		composeProjectLabel: "openclaw",
		composeConfigLabel:  composePath,
	}}
	sandbox := container.Summary{Labels: map[string]string{
		composeProjectLabel: "sandbox",
		composeConfigLabel:  composePath,
	}}
	orders := map[string][]container.Summary{
		"openclaw first": {openclaw, sandbox},
		"sandbox first":  {sandbox, openclaw},
	}
	for order, list := range orders {
		t.Run(order, func(t *testing.T) {
			// The display name normalizes to one of the projects.
			got, err := loadComposeProjectNameForLog(list, "OpenClaw", composePath)
			if err != nil || got != "openclaw" {
				t.Fatalf("display name: got %q, %v, want %q", got, err, "openclaw")
			}
			// An exact project name selects that project.
			got, err = loadComposeProjectNameForLog(list, "sandbox", composePath)
			if err != nil || got != "sandbox" {
				t.Fatalf("project name: got %q, %v, want %q", got, err, "sandbox")
			}
			// A name that matches neither project is ambiguous.
			got, err = loadComposeProjectNameForLog(list, "My App", composePath)
			if err == nil {
				t.Fatalf("ambiguous name: got %q, want an error", got)
			}
		})
	}
}
