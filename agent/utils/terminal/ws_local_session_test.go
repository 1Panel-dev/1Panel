package terminal

import "testing"

func TestBuildAIPastePayloadWrapsGeneratedCommand(t *testing.T) {
	got := buildAIPastePayload("ls -la /tmp")

	want := append([]byte{lineClearControl}, []byte("\x1b[200~ls -la /tmp\x1b[201~")...)
	if string(got) != string(want) {
		t.Fatalf("buildAIPastePayload() = %q, want %q", got, want)
	}
}

func TestBuildAIPastePayloadTrimsWhitespace(t *testing.T) {
	got := buildAIPastePayload("  ls -la /tmp  ")

	want := append([]byte{lineClearControl}, []byte("\x1b[200~ls -la /tmp\x1b[201~")...)
	if string(got) != string(want) {
		t.Fatalf("buildAIPastePayload() = %q, want %q", got, want)
	}
}
