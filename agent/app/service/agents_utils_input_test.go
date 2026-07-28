package service

import (
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

func TestBuildOpenclawModelEntryInputModalities(t *testing.T) {
	cases := []struct {
		name      string
		modelID   string
		model     dto.AgentAccountModel
		wantInput []string
	}{
		{
			name:      "minimax m3 accepts text image and video",
			modelID:   "MiniMax-M3",
			model:     dto.AgentAccountModel{ID: "MiniMax-M3", Name: "MiniMax M3"},
			wantInput: []string{"text", "image", "video"},
		},
		{
			name:      "minimax m3 lowercased accepts video",
			modelID:   "minimax-m3",
			model:     dto.AgentAccountModel{ID: "minimax-m3"},
			wantInput: []string{"text", "image", "video"},
		},
		{
			name:      "minimax m2.7 stays text only",
			modelID:   "MiniMax-M2.7",
			model:     dto.AgentAccountModel{ID: "MiniMax-M2.7", Name: "MiniMax M2.7"},
			wantInput: nil,
		},
		{
			name:      "claude sonnet accepts text and image",
			modelID:   "claude-sonnet-4-5",
			model:     dto.AgentAccountModel{ID: "claude-sonnet-4-5"},
			wantInput: []string{"text", "image"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry := buildOpenclawModelEntry(tc.modelID, tc.model)
			if len(entry.Input) != len(tc.wantInput) {
				t.Fatalf("input length: got %v, want %v", entry.Input, tc.wantInput)
			}
			for i, want := range tc.wantInput {
				if entry.Input[i] != want {
					t.Fatalf("input[%d]: got %q, want %q", i, entry.Input[i], want)
				}
			}
		})
	}
}
