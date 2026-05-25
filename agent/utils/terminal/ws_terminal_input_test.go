package terminal

import (
	"bytes"
	"testing"
)

func TestBuildAIPastePayloadAllowsNormalShellSyntax(t *testing.T) {
	payload, err := buildAIPastePayload("printf 'a' | grep a && echo done")
	if err != nil {
		t.Fatalf("buildAIPastePayload() error = %v, want nil", err)
	}
	if !bytes.Contains(payload, []byte("printf 'a' | grep a && echo done")) {
		t.Fatalf("buildAIPastePayload() = %q, want command in payload", payload)
	}
}

func TestBuildAIPastePayloadRejectsControlCharacters(t *testing.T) {
	for _, payload := range []string{
		"echo safe\nrm -rf /",
		"echo safe\rid",
		"echo\x00x",
		"echo\x1b[31m",
	} {
		if _, err := buildAIPastePayload(payload); err == nil {
			t.Fatalf("buildAIPastePayload(%q) error = nil, want rejection", payload)
		}
	}
}
