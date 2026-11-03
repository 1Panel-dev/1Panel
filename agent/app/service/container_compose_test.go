package service

import (
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestLoadComposeProjectNameForLog(t *testing.T) {
	list := []container.Summary{
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
			if got := loadComposeProjectNameForLog(list, tt.inputName, tt.composePath); got != tt.want {
				t.Fatalf("loadComposeProjectNameForLog(%q, %q) = %q, want %q", tt.inputName, tt.composePath, got, tt.want)
			}
		})
	}
}
